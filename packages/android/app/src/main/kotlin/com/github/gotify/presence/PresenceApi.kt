package com.github.gotify.presence

import com.github.gotify.client.model.Presence
import retrofit2.Call
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.PUT

internal interface PresenceApi {
    @GET("presence")
    fun list(): Call<Array<Presence>>

    @PUT("presence")
    fun put(@Body body: PresenceBody): Call<Presence>
}

internal data class PresenceBody(
    val state: String,
    val source: String,
    val detail: String?,
    val validForSeconds: Int
)
