# Protocolo de comunicação entre agentes

Todo agente do harness (Task Creator, Coder, Tester, PR Reviewer) se comunica
**exclusivamente** por mensagens num canal estilo Slack — um por empresa. Este
documento define o envelope, os tipos de mensagem e as regras de handoff da
esteira de tasks. É a fonte de verdade para qualquer runner de agente e espelha
`internal/core/domain.Message` + `internal/orchestrator/statemachine.go`.

## Envelope

Toda mensagem é um objeto JSON com exatamente estes campos:

| Campo | Tipo | Descrição |
|---|---|---|
| `type` | `"request" \| "response" \| "status"` | classe da mensagem |
| `from_agent` | string | nome do agente emissor |
| `to_agent` | string \| null | destinatário; `null` em `status`/broadcast |
| `in_reply_to` | string \| null | id da mensagem respondida; `null` se não for resposta |
| `requires_response` | bool | se `true`, o destinatário **deve** responder |
| `body` | string | texto livre, no tom do agente |
| `payload` | objeto | dados estruturados da etapa; nunca `null` (`{}` quando vazio) |

### Regras do envelope

- `type: "request"` → `to_agent` obrigatório; pode ou não exigir resposta.
- `type: "response"` → `to_agent` **e** `in_reply_to` obrigatórios; `requires_response` sempre `false`.
- `type: "status"` → `to_agent: null`, `in_reply_to: null`, `requires_response: false`. Evento que não bloqueia ninguém.
- Uma `request` com `requires_response: true` sem resposta dentro de N ciclos do orquestrador → a task vai para `needs_human`.
- Nunca dependa de parsear o `body`. O `payload` é o contrato estável; o `body` é humano.

## Máquina de estados da task

```
created → researching → queued → coding → testing → reviewing → { merged | needs_human | rejected }
                                    │          │
                    failed_build ◄──┘          └──► failed_test
                          │                              │
                          └──────► coding (retry) ◄──────┘
```

| Transição | Dono | Gatilho |
|---|---|---|
| `created → researching → queued` | Task Creator | pesquisa concluída, task priorizada com risco |
| `queued → coding` | Coder | pega a task, cria a branch |
| `coding → testing` | Coder | build + lint local passaram, código commitado |
| `coding → failed_build` | Coder | build/lint falhou |
| `testing → reviewing` | Tester | suíte passou (+ testes novos se faltava cobertura) |
| `testing → failed_test` | Tester | algum teste falhou |
| `failed_build / failed_test → coding` | Coder | nova tentativa (`retries++`) |
| `failed_* → needs_human` | orquestrador | `retries ≥ MaxRetries` |
| `reviewing → merged` | PR Reviewer | checklist ok **e** risco `low` |
| `reviewing → needs_human` | PR Reviewer | checklist ok mas risco `medium`/`high`, ou dúvida |
| `reviewing → rejected` | PR Reviewer | checklist reprovou de forma irrecuperável |

Limite de retries: **3** por task (`MaxRetries`, configurável). Na falha que
estoura o limite o agente para, emite `status` explicando o padrão da falha, e a
task vai para `needs_human` com o log anexado.

## Handoffs (payload por etapa)

### Task Creator → fila + Coder

O Task Creator cria a task via `POST /api/v1/tasks`, emite um `status`
(`event: "task_created"`), e então atribui ao Coder:

```json
{ "type": "request", "from_agent": "<tc>", "to_agent": "<coder>", "in_reply_to": null,
  "requires_response": true, "body": "...",
  "payload": { "task_id": "...", "title": "...", "description": "...",
               "acceptance": ["..."], "risk": "low", "branch": "autodev/<empresa>/<slug>" } }
```

O Coder responde confirmando: `payload: { "task_id": "...", "accepted": true, "branch": "..." }`.

### Coder → Tester

```json
{ "type": "request", "from_agent": "<coder>", "to_agent": "<tester>", "in_reply_to": null,
  "requires_response": true, "body": "...",
  "payload": { "task_id": "...", "branch": "...", "files_changed": ["..."],
               "needs_new_tests": true, "notes": "..." } }
```

### Tester → Coder (falha) · Tester → PR Reviewer (aprovado)

Falha: `response` à request do Coder com
`payload: { "task_id": "...", "passed": false, "failures": ["..."] }`.

Aprovado: nova `request` ao PR Reviewer com
`payload: { "task_id": "...", "branch": "...", "coverage_delta": "+2.1%", "new_tests": ["..."] }`.

### PR Reviewer → decisão

```json
{ "type": "status", "from_agent": "<pr>", "to_agent": null, "in_reply_to": null,
  "requires_response": false, "body": "...",
  "payload": { "task_id": "...", "pr_number": 42,
               "decision": "auto_merge" | "needs_human" | "rejected",
               "checklist": { "<item>": "pass" | "fail" | "n/a" } } }
```

## Status intermediários

Eventos que não bloqueiam ninguém — `payload: { "task_id": "...", "event": "<evt>" }`:
`branch_created`, `build_passed`, `build_failed`, `fix_retry`, `tests_generated`,
`pr_opened`, `pr_updated`, `task_created`.

## Idempotência

Cada agente, ao ser retomado, olha o estado salvo da task e a última mensagem do
canal referente a ela. Nunca recomeça do zero, nunca duplica branch/PR/mensagem.
