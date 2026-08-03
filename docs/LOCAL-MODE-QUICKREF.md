# AgentFlow Local Mode — Referencia rápida

Chuleta de bolsillo para el flujo de trabajo con Controller/Coder en modo
local. Para el detalle completo, ver `docs/LOCAL-MODE-GUIDE.md`.

## Secuencia de pasos

| # | Etapa | Qué le dices al Controller | Qué le dices al Coder |
|---|---|---|---|
| 1 | Conversación libre (backlog, ideas) | Lo que sea — sin comando especial, es chat normal | — |
| 2 | Decidiste qué hacer | `/agentflow-local-init` | — |
| 3 | Etapa A: análisis/impacto/plan | Responde sus preguntas sobre el problema; al final revisa `PLAN.md` y di **"aprobado"** o pide cambios | — |
| 4 | Etapa B: crea worktree + primera tarea | (nada — lo hace solo tras tu aprobación) | — |
| 5 | Vincular al Coder | Te dice a ti que lo hagas | Abre una **segunda sesión**, en el mismo worktree: `/agentflow-local-coder-init` |
| 6 | Coder ejecuta la tarea | (nada, esperas) | (nada — ya tiene `task.md`, trabaja solo: investiga, programa, prueba, commitea, escribe `result.md`) |
| 7 | Coder termina | — | Te avisa que terminó (tú lees `result.md` o se lo dices al Controller) |
| 8 | Revisar resultado | Vuelve a esa sesión (o abre una nueva: `/agentflow-local-resume`) y dile **"revisa el resultado"** | — |
| 9 | Aprobar o pedir corrección | Di **"aprobado, siguiente tarea"** o **"esto está mal, pídele que corrija X"** | — |
| 10 | Siguiente tarea | (nada — el Controller escribe la nueva `task.md` solo) | Abre **otra sesión nueva**: `/agentflow-local-coder-resume` |
| — repite 6→10 hasta terminar el plan — |
| 11 | Interrupción a medio camino (opcional) | `/agentflow-local-pause` en la sesión que necesitas cortar | igual, `/agentflow-local-pause` si es la sesión del Coder |
| 12 | Plan completo | El Controller te avisa que ya se puede mergear — di **"mergea y cierra"** | — |
| 13 | Cierre | (tú o el Controller corren `agentflow close`) | — |

## Reglas de oro

- **Coder**: siempre es sesión nueva por tarea — nunca "continúa" la sesión
  anterior. Primera vez → `/agentflow-local-coder-init`; todas las
  siguientes → `/agentflow-local-coder-resume`.
- **Controller**: una sola sesión larga mientras no haya un "reinicio"
  (aprobaste un plan, cerraste una deliberación, cambiaste el objetivo, o la
  conversación ya está muy larga) — ahí usas `/agentflow-local-resume` en
  una sesión nueva.
- **Antes de decidir el trabajo**: no hace falta ningún comando — es
  conversación libre con el Controller. Solo cuando ya decidiste qué hacer,
  invocas `/agentflow-local-init`.
- **`agentflow local status`**: en cualquier momento, para ver en qué va
  todo, incluyendo el progreso del plan (`N/M done`).
