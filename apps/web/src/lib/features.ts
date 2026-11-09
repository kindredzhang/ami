/**
 * Deployment feature switches.
 *
 * These are not user preferences — they describe what a given installation
 * ships. The Dongfeng intranet build runs a deliberately smaller product than
 * upstream:
 *
 *  - One bot per user. The platform's model is "one bot = one agent with its own
 *    VM", so 专家 (expert) is not a thing you manage here — the user has exactly
 *    one agent (东风助手) and everything is configured on it. Managing experts,
 *    creating experts, and the expert list all go away; what used to be called
 *    专家设置 becomes 智能体设置.
 *  - No external agents (ACP): no Codex / Claude Code / Grok / custom ACP.
 *  - No app market: the curated 专家-技能 lineup replaces browsing a market.
 *  - No channels/platform bindings (Feishu, WeChat, …) for now.
 *
 * The runtime code for all of it still exists in the repository; every
 * user-facing surface is switched off here rather than sprinkled as ad-hoc
 * `v-if`s. Flipping a flag back to true restores its surfaces — the underlying
 * APIs and components never left.
 */

/**
 * External agents (ACP): the composer's agent picker, the per-expert Agent tab,
 * the default-agent setting, the agent step in the create-expert flow, and the
 * agent badge on historical sessions.
 */
export const externalAgentsEnabled = false

/**
 * Multi-expert management: the 专家 settings page, its sidebar entry, and every
 * "新建专家" / "导入专家" affordance. With one bot per user there is nothing to
 * list or create — the single agent is reached through 智能体设置.
 */
export const expertManagementEnabled = false

/**
 * Channel bindings (the 平台 tab): Feishu / WeChat / webhook delivery for an
 * agent's messages. Not used yet; kept hidden rather than removed.
 */
export const channelsEnabled = false

/**
 * Per-agent app installs (the 应用 tab). The default app market does not exist
 * in this deployment; runtimes and toolchains are baked into the agent's VM
 * image at creation time instead of being installed on demand.
 */
export const agentAppsEnabled = false
