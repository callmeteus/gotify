package com.github.gotify.presence

import com.github.gotify.Settings
import com.github.gotify.api.Callback
import com.github.gotify.api.ClientFactory
import com.github.gotify.client.model.Presence
import org.tinylog.kotlin.Logger

/** Publishes presence to the server and reads the states of the other users. */
internal class PresenceReporter(private val settings: Settings, private val store: PresenceStore) {

    private fun api(): PresenceApi =
        ClientFactory.clientToken(settings).createService(PresenceApi::class.java)

    fun publish(decision: Decision, onResult: (Boolean) -> Unit = {}) {
        api().put(
            PresenceBody(
                state = decision.state.wire,
                source = decision.source.wire,
                detail = decision.detail,
                validForSeconds = store.presenceTtlSeconds
            )
        ).enqueue(
            Callback.call(
                onSuccess = { onResult(true) },
                onError = { error ->
                    Logger.error(error, "presence publish failed")
                    onResult(false)
                }
            )
        )
    }

    fun refresh(onResult: (Array<Presence>) -> Unit = {}) {
        api().list().enqueue(
            Callback.call(
                onSuccess = Callback.SuccessBody { presences -> onResult(presences) },
                onError = { error -> Logger.error(error, "presence list failed") }
            )
        )
    }
}
