package com.iskenkenya.commory.mobile.agent

import com.iskenkenya.commory.mobile.runtime.RuntimeMode
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class AgentPrivacyPolicyTest {
    @Test
    fun localModeKeepsAgentContextLocal() {
        assertEquals(AgentProviderPolicy.LOCAL_ONLY, AgentPrivacyPolicy.policyFor(RuntimeMode.LOCAL_ONLY))
        assertEquals(AgentProviderPolicy.SERVER_ALLOWED, AgentPrivacyPolicy.policyFor(RuntimeMode.COMMORY_SERVER))
    }

    @Test
    fun previewsAreTrimmedBeforeProviderUse() {
        val preview = AgentPrivacyPolicy.trimPreview("abcdef", maxChars = 3)
        assertEquals("abc", preview)
        assertTrue(preview.length <= 3)
    }
}
