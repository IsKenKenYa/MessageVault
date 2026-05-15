package imken.messagevault.mobile.agent

import imken.messagevault.mobile.runtime.RuntimeMode

enum class AgentContextSource {
    MESSAGE,
    CALL_LOG,
    CONTACT
}

enum class AgentProviderPolicy {
    LOCAL_ONLY,
    SERVER_ALLOWED
}

data class AgentContextReference(
    val source: AgentContextSource,
    val sourceId: String,
    val participantIds: List<String>,
    val timestampMillis: Long,
    val preview: String
)

data class AgentContextWindow(
    val query: String,
    val references: List<AgentContextReference>,
    val maxItems: Int = 20,
    val maxPreviewChars: Int = 240,
    val policy: AgentProviderPolicy
)

object AgentPrivacyPolicy {
    fun policyFor(runtimeMode: RuntimeMode): AgentProviderPolicy {
        return when (runtimeMode) {
            RuntimeMode.LOCAL_ONLY -> AgentProviderPolicy.LOCAL_ONLY
            RuntimeMode.COMMORY_SERVER -> AgentProviderPolicy.SERVER_ALLOWED
        }
    }

    fun trimPreview(value: String, maxChars: Int = 240): String {
        return if (value.length <= maxChars) value else value.take(maxChars)
    }
}
