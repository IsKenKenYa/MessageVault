package imken.messagevault.sdk.backup.msglayer.model

const val MSG_LAYER_VERSION = "msglayer/v0.1"

data class MsgLayerRootExport(
    val version: String = MSG_LAYER_VERSION,
    val exportedAt: String,
    val source: MsgLayerSource,
    val identities: List<MsgLayerIdentity>,
    val events: List<MsgLayerEvent>,
    val indexes: Map<String, Any?>? = null
)

data class MsgLayerSource(
    val platform: String,
    val deviceId: String,
    val appVersion: String
)

data class MsgLayerIdentity(
    val id: String,
    val type: String,
    val displayName: String,
    val phones: List<String>,
    val emails: List<String>,
    val avatar: String? = null,
    val labels: List<String> = emptyList(),
    val meta: Map<String, Any?> = emptyMap()
)

data class MsgLayerEvent(
    val id: String,
    val type: String,
    val timestamp: String,
    val direction: String,
    val participants: List<String>,
    val content: Map<String, Any?>,
    val meta: Map<String, Any?> = emptyMap(),
    val relations: List<MsgLayerRelation> = emptyList()
)

data class MsgLayerRelation(
    val type: String,
    val target: String
)
