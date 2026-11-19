package com.github.gotify.presence

import com.github.gotify.alarm.IncidentStages
import com.github.gotify.client.JSON
import com.github.gotify.client.model.Incident
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class IncidentStagesTest {
    private fun incident(state: String, currentStage: Int): Incident =
        JSON().gson.fromJson(
            """{"id":1,"appid":2,"state":"$state","currentStage":$currentStage,"stages":[{"label":"I'm up"},{"label":"I'm at my computer"}]}""",
            Incident::class.java
        )

    @Test
    fun ringingStartsAtStageZero() = assertEquals(0, IncidentStages.next(incident("ringing", -1)))

    @Test
    fun waitingAdvancesToFollowingStage() = assertEquals(1, IncidentStages.next(incident("waiting", 0)))

    @Test
    fun closedHasNoStage() = assertNull(IncidentStages.next(incident("closed", 1)))

    @Test
    fun missingIncidentHasNoStage() = assertNull(IncidentStages.next(null))

    @Test
    fun outOfRangeStageIsRejected() = assertNull(IncidentStages.next(incident("waiting", 1)))
}
