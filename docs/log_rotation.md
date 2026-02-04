# Política de Rotação de Logs

**Versão**: 1.0.0  
**Data**: 2026-02-03  
**Escopo**: Ambientes de produção sem agregação centralizada

---

## 1. Contexto Técnico

### Comportamento da Aplicação
A aplicação:
- Loga para `stdout` e `stderr` (padrão POSIX)
- Usa logs estruturados em JSON via `log/slog`
- **NÃO** gerencia rotação internamente
- **NÃO** escreve em arquivos diretamente

### Responsabilidade do Host
Toda rotação de logs é **responsabilidade do sistema operacional** ou do runtime de containers (Docker, systemd, etc.).

---

## 2. Política Mínima Recomendada

### Parâmetros de Rotação

| Parâmetro | Valor Recomendado | Justificativa |
|-----------|-------------------|---------------|
| Tamanho máximo por arquivo | 100 MB | Facilita análise manual, evita arquivos gigantes |
| Número de arquivos mantidos | 7 | Uma semana de histórico para troubleshooting |
| Período de retenção | 7 dias | Balanceamento entre espaço e rastreabilidade |
| Compressão | Sim (gzip) | Reduz uso de disco em ~70% |
| Rotação por tempo | Diária (meia-noite) | Alinhamento com ciclos de backup |

### Espaço em Disco Estimado
- **Sem compressão**: 7 × 100 MB = 700 MB
- **Com compressão**: ~210 MB
- **Margem de segurança**: 1 GB total

---

## 3. Configuração via logrotate

### Para Deployment Tradicional (Systemd)

Se a aplicação roda via `systemd` e loga para arquivo via redirecionamento:

```bash
# Serviço systemd (/etc/systemd/system/doligo.service)
[Service]
ExecStart=/usr/local/bin/doligo
StandardOutput=append:/var/log/doligo/app.log
StandardError=append:/var/log/doligo/error.log
```

Criar arquivo de configuração do `logrotate`:

```bash
# /etc/logrotate.d/doligo
/var/log/doligo/*.log {
    daily
    rotate 7
    size 100M
    compress
    delaycompress
    missingok
    notifempty
    create 0640 doligo doligo
    sharedscripts
    postrotate
        systemctl reload doligo.service > /dev/null 2>&1 || true
    endscript
}
```

**Validar configuração**:
```bash
# Testar sintaxe
sudo logrotate -d /etc/logrotate.d/doligo

# Forçar rotação manual (debug)
sudo logrotate -f /etc/logrotate.d/doligo
```

### Para Deployment sem Systemd

Se a aplicação roda diretamente e loga para arquivo:

```bash
# Iniciar aplicação com redirecionamento
nohup /usr/local/bin/doligo >> /var/log/doligo.log 2>&1 &
```

Configuração `logrotate`:

```bash
# /etc/logrotate.d/doligo
/var/log/doligo.log {
    daily
    rotate 7
    size 100M
    compress
    delaycompress
    missingok
    notifempty
    create 0640 root root
    copytruncate
}
```

**Nota**: `copytruncate` é necessário pois a aplicação não reabre o arquivo após rotação.

---

## 4. Configuração via journald (Systemd)

### Para Deployment via Systemd (Recomendado)

Se a aplicação roda via `systemd`, os logs vão automaticamente para `journald`:

```bash
# Serviço systemd (/etc/systemd/system/doligo.service)
[Service]
ExecStart=/usr/local/bin/doligo
# Logs vão para journald automaticamente
```

Configurar limites do `journald`:

```bash
# /etc/systemd/journald.conf
[Journal]
SystemMaxUse=1G
SystemKeepFree=2G
SystemMaxFileSize=100M
MaxRetentionSec=7day
Compress=yes
```

**Aplicar configuração**:
```bash
sudo systemctl restart systemd-journald
```

**Visualizar logs**:

```bash
# Logs da aplicação
journalctl -u doligo.service -f

# Logs com filtro JSON
journalctl -u doligo.service -o json-pretty

# Logs das últimas 24h
journalctl -u doligo.service --since "24 hours ago"
```

---

## 5. Configuração via Docker

### Docker Logging Driver (json-file)

Se a aplicação roda em container Docker:

```yaml
# docker-compose.yml
services:
  doligo:
    image: doligo:latest
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "7"
        compress: "true"
```

Ou via `docker run`:

```bash
docker run -d \
  --name doligo \
  --log-driver json-file \
  --log-opt max-size=100m \
  --log-opt max-file=7 \
  --log-opt compress=true \
  doligo:latest
```

**Visualizar logs**:

```bash
# Logs do container
docker logs -f doligo

# Logs com timestamps
docker logs -f --timestamps doligo
```

### Docker Logging Driver (journald)

Se o host usa `systemd`:

```yaml
# docker-compose.yml
services:
  doligo:
    image: doligo:latest
    logging:
      driver: "journald"
      options:
        tag: "doligo"
```

**Visualizar logs**:

```bash
# Logs via journalctl
journalctl CONTAINER_NAME=doligo -f
```

---

## 6. Validação da Configuração

### Checklist de Validação

- [ ] Política de rotação está configurada (logrotate, journald ou Docker)
- [ ] Tamanho máximo por arquivo definido
- [ ] Número de arquivos de retenção configurado
- [ ] Compressão está habilitada
- [ ] Espaço em disco disponível (mínimo 1GB)
- [ ] Rotação foi testada manualmente

### Comandos de Teste

#### logrotate
```bash
# Simular rotação (dry-run)
sudo logrotate -d /etc/logrotate.d/doligo

# Forçar rotação
sudo logrotate -f /etc/logrotate.d/doligo

# Verificar arquivos rotacionados
ls -lh /var/log/doligo/
```

#### journald
```bash
# Verificar uso de disco
journalctl --disk-usage

# Limpar logs antigos manualmente
sudo journalctl --vacuum-time=7d
sudo journalctl --vacuum-size=500M
```

#### Docker
```bash
# Verificar configuração de logging
docker inspect doligo | grep -A 10 LogConfig

# Verificar tamanho dos logs
du -sh /var/lib/docker/containers/$(docker ps -qf name=doligo)/*-json.log*
```

---

## 7. Monitoramento e Alertas (Básico)

### Indicadores de Saúde

Monitorar manualmente:
- Uso de disco em `/var/log` ou `/var/lib/docker`
- Quantidade de arquivos `.log` ou `.log.gz`
- Idade do arquivo mais antigo

### Exemplo de Script de Monitoramento

```bash
#!/bin/bash
# /usr/local/bin/check-log-disk.sh

LOG_DIR="/var/log/doligo"
MAX_USAGE_PERCENT=80

USAGE=$(df -h "$LOG_DIR" | awk 'NR==2 {print $5}' | sed 's/%//')

if [ "$USAGE" -gt "$MAX_USAGE_PERCENT" ]; then
    echo "ALERTA: Uso de disco em $LOG_DIR está em ${USAGE}%"
    exit 1
else
    echo "OK: Uso de disco em ${USAGE}%"
    exit 0
fi
```

**Nota**: Este script é apenas exemplo. Automação de alertas está fora do escopo.

---

## 8. Limitações e Riscos Conhecidos

### Limitações da Estratégia
- Rotação depende de configuração correta do host
- Sem validação automática de configuração
- Sem agregação centralizada de logs
- Sem alertas automáticos de falha de rotação

### Riscos Aceitos
- Logs podem ser perdidos se rotação falhar silenciosamente
- Disco pode encher se política não for aplicada
- Troubleshooting limitado a 7 dias de histórico

### Quando Revisar Esta Estratégia
- Volume de logs ultrapassar 1GB/dia
- Necessidade de retenção > 7 dias
- Necessidade de busca full-text em logs
- Múltiplas instâncias da aplicação (agregação necessária)

---

## 9. Próximos Passos (Fora de Escopo)

Melhorias futuras não implementadas nesta fase:
- Agregação centralizada (ELK, Loki, CloudWatch)
- Alertas automáticos de falha de rotação
- Exportação de logs para storage externo
- Análise de logs com ferramentas de BI
- Correlação de logs entre múltiplos serviços

---

**Documento produzido em conformidade com FASE 12.5**  
**Nenhuma implementação de rotação interna foi criada**  
**Configuração é responsabilidade do host/runtime**
