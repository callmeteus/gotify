package com.github.gotify.presence

import com.github.gotify.client.JSON
import com.github.gotify.client.model.Message
import com.github.gotify.client.model.PresenceFrame
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

/** Locks the WebSocket wire contract added for presence on the Android client. */
class PresenceFrameTest {
    private val gson = JSON().gson

    @Test
    fun parsesPresenceFrame() {
        val json =
            """{"type":"presence","data":{"userId":2,"state":"at_desk","source":"wifi","detail":"office","reportedAt":"2020-01-01T00:00:00Z","validUntil":"2020-01-01T00:30:00Z"}}"""
        val frame = gson.fromJson(json, PresenceFrame::class.java)

        assertTrue(frame.isPresence())
        assertEquals("presence", frame.type)
        assertEquals(2L, frame.data!!.userId)
        assertEquals("at_desk", frame.data!!.state)
        assertEquals("wifi", frame.data!!.source)
        assertEquals("office", frame.data!!.detail)
        assertEquals(
            java.time.OffsetDateTime.parse("2020-01-01T00:30:00Z"),
            frame.data!!.validUntil
        )
    }

    @Test
    fun messageFrameIsNotMistakenForPresence() {
        val json =
            """{"id":7,"appid":3,"message":"disk full","title":"backup","date":"2020-01-01T00:00:00Z","priority":8}"""
        val frame = gson.fromJson(json, PresenceFrame::class.java)

        assertFalse(frame.isPresence())
        val message = gson.fromJson(json, Message::class.java)
        assertEquals(7L, message.id)
        assertEquals("disk full", message.message)
        assertNull(message.incident)
    }

    @Test
    fun unknownStateFallsBackToUnknown() {
        assertEquals(PresenceState.AT_DESK, PresenceState.fromWire("at_desk"))
        assertEquals(PresenceState.AWAY, PresenceState.fromWire("away"))
        assertEquals(PresenceState.UNKNOWN, PresenceState.fromWire("busy"))
        assertEquals(PresenceState.UNKNOWN, PresenceState.fromWire(null))
    }
}
