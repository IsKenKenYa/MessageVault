package com.iskenkenya.commory.sdk.backup.msglayer

import com.google.gson.FieldNamingPolicy
import com.google.gson.GsonBuilder
import com.iskenkenya.commory.sdk.backup.msglayer.model.MsgLayerRootExport

class MsgLayerSerializer {

    private val gson = GsonBuilder()
        .setFieldNamingPolicy(FieldNamingPolicy.LOWER_CASE_WITH_UNDERSCORES)
        .serializeNulls()
        .disableHtmlEscaping()
        .setPrettyPrinting()
        .create()

    fun toJson(export: MsgLayerRootExport): String = gson.toJson(export)

    fun gson() = gson
}
