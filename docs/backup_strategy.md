# Estratégia de Backup de Banco de Dados

**Versão**: 1.0.0  
**Data**: 2026-02-03  
**Escopo**: Ambiente único, banco único, volume pequeno/médio

---

## 1. Contexto e Premissas

### Escopo da Estratégia
- **Tipo de Backup**: Full (completo)
- **Banco Suportado**: PostgreSQL E MySQL
- **Ambiente**: Produção single-instance
- **Volume Estimado**: Até 100GB (ajustável)

### O que ESTÁ Coberto
- Backup de schema e dados
- Retenção de versões históricas
- Procedimento de restauração manual

### O que NÃO Está Coberto
- Backup incremental/diferencial
- Replicação em tempo real
- Multi-região ou disaster recovery
- Automação via CI/CD
- Monitoramento de sucesso de backup

---

## 2. Estratégia Definida

### Tipo de Backup
**Full Backup** (backup completo de schema + dados)

**Justificativa**:
- Simplicidade operacional
- Recuperação previsível
- Adequado para volumes não massivos
- Compatível com ferramentas nativas

### Frequência Recomendada
- **Diária**: 1x por dia (00:00 UTC ou horário de menor carga)
- **Semanal** (opcional): Backup adicional aos domingos para retenção longa

### Retenção Sugerida
- **Backups diários**: 7 dias (última semana)
- **Backups semanais**: 4 semanas (último mês)
- **Total de armazenamento**: ~11 backups simultâneos

### Local de Armazenamento
- **Primário**: Sistema de arquivos local seguro (fora do diretório da aplicação)
- **Recomendado**: Volume persistente montado (ex: `/backup` ou `/mnt/backup`)
- **Ideal** (fora de escopo desta fase): Cópia para storage externo (S3, NFS, etc.)

---

## 3. Procedimentos PostgreSQL

### Backup Completo

```bash
# Variáveis de ambiente (definir antes)
export PGHOST=localhost
export PGPORT=5432
export PGDATABASE=doligo_db
export PGUSER=doligo_user
export PGPASSWORD=senha_segura

# Executar backup
BACKUP_FILE="/backup/doligo_$(date +%Y%m%d_%H%M%S).sql.gz"
pg_dump --format=plain --no-owner --no-acl | gzip > "$BACKUP_FILE"

# Verificar sucesso
if [ $? -eq 0 ]; then
    echo "Backup criado: $BACKUP_FILE"
else
    echo "ERRO: Falha no backup"
    exit 1
fi
```

### Restauração

```bash
# Identificar arquivo de backup
BACKUP_FILE="/backup/doligo_20260203_000000.sql.gz"

# ATENÇÃO: Este comando SOBRESCREVE o banco existente
# Certifique-se de ter backup do estado atual antes de restaurar

# Restaurar
gunzip -c "$BACKUP_FILE" | psql -d doligo_db

# Verificar integridade
psql -d doligo_db -c "SELECT COUNT(*) FROM users;"
```

### Rotação de Backups (Manual)

```bash
# Manter apenas últimos 7 dias
find /backup -name "doligo_*.sql.gz" -mtime +7 -delete
```

---

## 4. Procedimentos MySQL

### Backup Completo

```bash
# Variáveis de ambiente (definir antes)
export MYSQL_HOST=localhost
export MYSQL_PORT=3306
export MYSQL_DATABASE=doligo_db
export MYSQL_USER=doligo_user
export MYSQL_PASSWORD=senha_segura

# Executar backup
BACKUP_FILE="/backup/doligo_$(date +%Y%m%d_%H%M%S).sql.gz"
mysqldump --single-transaction --routines --triggers --events \
    -h "$MYSQL_HOST" -P "$MYSQL_PORT" \
    -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" \
    "$MYSQL_DATABASE" | gzip > "$BACKUP_FILE"

# Verificar sucesso
if [ $? -eq 0 ]; then
    echo "Backup criado: $BACKUP_FILE"
else
    echo "ERRO: Falha no backup"
    exit 1
fi
```

### Restauração

```bash
# Identificar arquivo de backup
BACKUP_FILE="/backup/doligo_20260203_000000.sql.gz"

# ATENÇÃO: Este comando SOBRESCREVE o banco existente
# Certifique-se de ter backup do estado atual antes de restaurar

# Restaurar
gunzip -c "$BACKUP_FILE" | mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" \
    -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE"

# Verificar integridade
mysql -h "$MYSQL_HOST" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" \
    -e "SELECT COUNT(*) FROM users;" "$MYSQL_DATABASE"
```

### Rotação de Backups (Manual)

```bash
# Manter apenas últimos 7 dias
find /backup -name "doligo_*.sql.gz" -mtime +7 -delete
```

---

## 5. Checklist de Execução

### Antes do Primeiro Backup
- [ ] Criar diretório `/backup` com permissões adequadas (`chmod 700`)
- [ ] Verificar espaço em disco disponível (mínimo 2x tamanho do banco)
- [ ] Testar credenciais de acesso ao banco
- [ ] Executar backup de teste e validar restauração

### Rotina Diária
- [ ] Executar comando de backup conforme procedimento
- [ ] Verificar criação do arquivo `.sql.gz`
- [ ] Validar tamanho do arquivo (não deve ser 0 bytes)
- [ ] Registrar sucesso/falha em log operacional

### Rotina Semanal
- [ ] Executar rotação de backups antigos
- [ ] Verificar espaço em disco restante
- [ ] Testar restauração de um backup aleatório (validação)

---

## 6. Recuperação de Desastres (Básico)

### Cenário 1: Corrupção de Dados Recente
1. Identificar backup mais recente anterior à corrupção
2. Parar a aplicação (`systemctl stop doligo` ou equivalente)
3. Executar restauração conforme procedimento
4. Validar integridade dos dados críticos
5. Reiniciar aplicação
6. Monitorar logs de erro

### Cenário 2: Perda Total do Banco
1. Recriar banco vazio com credenciais corretas
2. Executar migrações (se necessário): `golang-migrate up`
3. Executar restauração do backup mais recente
4. Validar schema e dados
5. Reiniciar aplicação

---

## 7. Limitações e Riscos Conhecidos

### Limitações da Estratégia
- **Ponto de recuperação**: Até 24h de perda de dados (backup diário)
- **Tempo de recuperação**: Proporcional ao tamanho do banco (não otimizado)
- **Sem automação**: Dependente de execução manual ou cron externo
- **Sem validação automática**: Integridade dos backups não é verificada automaticamente

### Riscos Aceitos
- Backup armazenado localmente (sem redundância geográfica)
- Falha de backup pode passar despercebida sem monitoramento
- Restauração não testada regularmente pode falhar quando necessária

### Quando Revisar Esta Estratégia
- Volume de dados ultrapassar 100GB
- RTO (Recovery Time Objective) exigir < 1 hora
- RPO (Recovery Point Objective) exigir < 24 horas
- Necessidade de multi-região ou disaster recovery

---

## 8. Próximos Passos (Fora de Escopo)

Melhorias futuras não implementadas nesta fase:
- Automação via `cron` ou `systemd.timer`
- Upload para storage externo (S3, GCS, Azure Blob)
- Notificação de sucesso/falha via e-mail ou webhook
- Validação automática de integridade de backups
- Backup incremental para redução de espaço

---

**Documento produzido em conformidade com FASE 12.3**  
**Nenhum script automático foi criado**  
**Nenhuma integração cloud foi implementada**
