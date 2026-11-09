# Experts

An **expert** is a named capability bundle a subagent run can execute as. It is
not a bot, not a second agent, and not a runtime switch: it is a narrower view
of the bot that owns it.

This document covers what an expert is, how a run resolves one, and the two
places the implementation has to be careful.

## Why not bots, and why not a runtime switch

The bot is a heavy aggregate: it owns a workspace container, an IM identity, an
ACL, a backup unit, an App install target, and a memory subject. Creating one
per role means every role pays for all of that, and "which role answered this"
becomes "which bot answered this" — a question the UI then has to answer by
navigating away from the conversation.

An expert narrows exactly two things:

- **which skills are visible** to the run, and
- **which persona files** the system prompt reads.

Everything else — tools, connectors, model, approval policy — stays whatever the
bot already grants. An expert therefore can never widen the bot's capability
surface, only narrow it. That property is what makes experts safe to create
freely; a second bot cannot make that claim.

## Storage

Experts live as files in the bot workspace, not in a table:

```
/data/experts/<name>/
  expert.json     # display_name, description, skills[]
  AGENTS.md       # persona
  SOUL.md         # persona (optional)
  IDENTITY.md     # persona (optional)
```

`expert.json` is optional. A directory with only markdown is a valid expert, so
a hand-written expert needs no ceremony. A directory with neither is skipped as
a stray.

Files rather than rows, for three reasons that all matter in practice:

- An expert can edit its own persona, because the files are in the workspace it
  already has.
- Experts travel with bot backup/export without a schema of their own.
- A user can author or review one with a text editor, exactly as they would a
  skill.

`name` must match `^[a-z0-9][a-z0-9_-]{0,63}$`: it becomes a directory, a URL
segment, and a token in the parent model's tool call, and a name with a slash or
a space is a name the model will mangle.

### `skills: []` means inherit, not "none"

An empty or absent `skills` list means **run with the bot's full effective skill
set**. A persona-only expert — one created to change voice, not capability —
must not be silently stripped of every skill.

This is why the resolver distinguishes two states rather than one:

| expert names skills | `Skills` | `SkillsScoped` | run uses |
|---|---|---|---|
| no | `nil` | `false` | the bot's full set |
| yes | non-nil (possibly empty) | `true` | exactly that map |

A non-nil **empty** map is a real outcome: an expert whose named skills have all
been uninstalled runs with no skills. Falling back to the full set there would
grant a run capabilities its author never intended, which is the one failure
mode this design exists to prevent.

## Resolution path

The parent model discovers experts with `list_experts` and names one on
`spawn_agent`:

```jsonc
{ "task": "count the rows", "expert": "ops-analyst" }
```

`SpawnProvider.resolveExpert` loads the definition and narrows `skills` against
the bot's effective set. Names the bot does not have are dropped and logged;
they are not an error, because one stale manifest entry should not make an
expert unusable for a reason the caller cannot act on.

An unknown expert name **fails before a subagent session exists**, and the error
tells the model to call `list_experts`. A bare "not found" makes the model guess
again, and a guessed expert silently runs with the wrong skills.

The resolved expert then travels as data:

- `SpawnExpert.Skills` / `SkillsScoped` → `SpawnRunConfig.Skills` and
  `SpawnIdentity.SkillsScoped`
- `SpawnExpert.SystemSection` → appended to the subagent system prompt
- `expert_name` → thread metadata, so a follow-up keeps the scope

### Follow-ups keep the expert

`send_message` has no `expert` argument, and it should not need one. The expert
name is stored on the created thread's metadata and read back into
`agentRecord.ExpertName`, so the second message to an agent keeps running as the
same expert rather than quietly reverting to the bot's full skill set.

## The two places to be careful

### 1. The skill provider reloads the bot's catalog

`SkillProvider.Tools` used to do this unconditionally:

```go
if p.load != nil {
    loaded, err := p.load(ctx, session.BotID)
    session.Skills = loaded   // overwrites whatever upstream decided
}
```

Its loader is authoritative for the bot — it re-reads every skill directory on
every call. That is correct for a normal session and wrong for an expert, where
upstream has already decided the set. Without a guard, an expert's slice comes
back as everything the bot has, and the feature silently does nothing.

The guard is the explicit `SkillsScoped` flag, not a heuristic. Tampering with
the loader's scope instead would break the cache-sharing the loader relies on.

### 2. The subagent prompt has no skills in it

`SpawnSystemPrompt` renders `GenerateSystemPrompt(SystemPromptParams{SessionType})`
— no skills, no files, no bot info. A subagent gets a skeleton prompt and has to
call `list_skills` to learn what exists.

This is why the expert's persona is **appended** to that prompt rather than
replacing it: the skeleton still carries the standing rules a run needs (tool-use
discipline, safety, how to report back). An expert refines *who* is doing the
work; it does not get to drop those rules.

The persona section is fenced under `## Active expert: <name>`. Without that
boundary, an expert's `AGENTS.md` reads as a replacement for the bot's
`AGENTS.md` rather than an addition to it.

## What a subagent cannot do

`spawn_agent` already refuses to run in a subagent, and `ask_user` and direct
message sending are gated on `IsSubagent`. For the "main agent delegates, expert
executes" split this is exactly right: the expert is a pure executor, and the
main agent owns the conversation with the user. No extra constraint is needed.

## Frontend

The web client shows expert work without leaving the conversation:

- `spawn_agent` with an `expert` argument renders as **派遣专家** rather than
  **派遣智能体**, with the expert name as the target — the expert is the stable
  half of the call while the task text varies every time.
- `expert-chip.vue` renders the identity token next to the run status.
- The spawned session's panel is titled `<expert> · <task>`, so a workspace with
  several side-by-side subagents says which expert each one is.
- `list_experts` renders as an expert list (name, description) rather than the
  "no tasks" empty state.

The expert is read from the call's **arguments**, not its result: a running
spawn has no result yet, and that is precisely when the reader wants to know who
is working.

## Adding an expert

```sh
# in the bot workspace
mkdir -p /data/experts/ops-analyst
cat > /data/experts/ops-analyst/expert.json <<'JSON'
{
  "display_name": "运营数据分析专家",
  "description": "分析运营看板数据，回答指标口径与趋势问题",
  "skills": ["officecli"]
}
JSON
cat > /data/experts/ops-analyst/AGENTS.md <<'MD'
# 运营数据分析专家

先给结论，再给依据。数据拿不到就说拿不到，不要估。
MD
```

The main agent can then `list_experts` and pick it by name. No restart, no
schema, no migration: the store reads the workspace on demand.

## Using an expert: the @ picker

The feature is only reachable through the composer, so this is the part that has
to match the product story rather than the storage story.

1. Type `@` anywhere in the composer (start of the draft, or after a space). The
   picker opens above the input with a **专家 / Experts** group listing every
   expert in the workspace, with its description and how many skills it scopes.
2. Pick one with the mouse, or `↑`/`↓` then `Enter`. The token is removed from
   the draft and the expert becomes a removable chip on the composer.
3. Type the task and send. The chip is per message: it is consumed by that send,
   so the next message goes back to the plain agent unless you `@` again.

What happens server-side on that send:

- `POST /bots/{bot_id}/experts` backs the picker and reads the same workspace
  `spawn_agent` resolves against, so anything listed can actually be given work.
- The WS `message` frame carries `expert: {name, display_name}`. Before the turn
  starts, the server resolves it; an unknown name fails the turn immediately
  with `expert.not_found` instead of letting the model improvise.
- For a valid expert the turn still runs on the **main agent**, which receives a
  directive instruction naming the expert and telling it to call
  `spawn_agent(expert="…")` with the user's request. That is deliberate: the
  expert's work has to stay a separate subagent run, with its own scoped skills
  and its own transcript, not a persona swap on the main agent.
- The subagent then runs with exactly the expert's skill slice (or the bot's
  whole set when the expert declares no skills) and the expert's persona
  appended to its system prompt.

### Why @ delegates instead of switching the main agent

Switching the main agent's persona for one message would produce a nicer-looking
single turn, but it loses the two things the expert concept exists for: the
scoped skill set would leak back to the parent on the next turn, and the work
would be indistinguishable from the main agent's own. Routing through
`spawn_agent` keeps the expert's messages, thinking and tool calls in their own
subagent panel, which is what the UI renders.

### What is not scoped yet

- **Connectors / MCP** stay bot-scoped. An expert narrows skills and persona
  only; per-expert connector credentials would need a nullable expert id on the
  binding table.
- **External runtimes** (codex / claudecode / ACP) read the filesystem directly
  and ignore `RunConfig.Skills`, so skill scoping does not apply to them.

## Demo experts: `docs`-backed definitions the ×4 catalog mirrors

The four experts the 专家-技能 catalog shows are also real experts under
`/data/experts/`, so typing `@` lists the same names the catalog advertises.

| `name` | `display_name` | skills in manifest |
|---|---|---|
| `zhiying-ops` | 智算运营助理 | omitted → inherits all |
| `compliance-ops` | 智驾合规运营小助理 | omitted → inherits all |
| `scenario-ops` | 500+场景运营经理 | omitted → inherits all |
| `qt-platform-ops` | 擎天AI中台运营专家 | `skill-creator` (scoped) |

**Why three omit `skills`.** The catalog names skills (`xlsx`, `pdf`, `docx`, …)
that this bot does not have; `Store.load` narrows a manifest's `skills` against
the bot's effective set, so naming them would have produced an *empty scoped*
set — an expert with no skills at all. `nil` means "inherit everything", so
omitting the key is the correct expression of "this expert has no skill opinion".
Only `qt-platform-ops` claims a skill the bot actually has, which makes it the
one expert that demonstrates narrowing (`list_skills` inside it returns exactly
`skill-creator`).

## Two failure modes worth remembering

**A model row without `tool_call: true` silently removes every tool.** Whether a
turn gets tools at all is gated on `chatModel.HasCompatibility(models.CompatToolCall)`
(`internal/agent/application/service.go`). A model added by hand through
`POST /models` starts with `config = {"schema_version": 2}` — no `tool_call` — so
`assembleTools` never runs and the model truthfully reports "no tools available".
It looks like a subagent bug and is not one. Rows imported from a provider that
declares capabilities are fine; hand-added rows need the compat bit copied.

**`reasoning.efforts` must match what the upstream accepts.** DeepSeek V4.1
rejects `medium` outright: *"reasoning_effort must be low, high, xhigh, max, or an
integer within [1, 100]"*. A row advertising `[low, medium, high]` makes every
turn fail with `agent.provider_request_rejected`, which reads as a provider outage.
Record the upstream's real menu in `models.config.reasoning.efforts`.
