package com.github.gotify.alarm

import com.github.gotify.client.model.Incident
import retrofit2.Call
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Path

internal interface IncidentApi {
    @GET("incident")
    fun list(): Call<Array<Incident>>

    @POST("incident/{id}/advance")
    fun advance(
        @Path("id") id: Long,
        @Body body: AdvanceStage
    ): Call<Incident>
}

internal data class AdvanceStage(val stage: Int)
