package com.github.gotify.alarm

import android.annotation.SuppressLint
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.graphics.Color
import android.os.Build
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.view.WindowManager
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import com.google.android.material.color.MaterialColors
import com.github.gotify.R
import com.github.gotify.Settings
import com.github.gotify.Utils
import com.github.gotify.api.Callback
import com.github.gotify.api.ClientFactory
import com.github.gotify.client.model.Incident
import com.github.gotify.client.model.Message
import com.github.gotify.databinding.ActivityAlarmBinding
import com.github.gotify.service.WebSocketService
import java.time.Duration
import java.time.LocalTime
import java.time.OffsetDateTime
import java.time.format.DateTimeFormatter

internal class AlarmActivity : AppCompatActivity() {
    private lateinit var binding: ActivityAlarmBinding
    private var message: Message? = null
    private var appName: String = ""
    private val clock = Handler(Looper.getMainLooper())
    private val tick = object : Runnable {
        override fun run() {
            render()
            clock.postDelayed(this, 1000)
        }
    }
    private val receiver = object : BroadcastReceiver() {
        override fun onReceive(context: Context, intent: Intent) {
            val raw = intent.getStringExtra("message") ?: return
            val incoming = Utils.JSON.fromJson(raw, Message::class.java)
            val incident = incoming.incident ?: return
            if (incident.kind == "expire") {
                finish()
                return
            }
            if (!shouldApply(message?.incident, incident)) {
                return
            }
            message = incoming
            render()
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O_MR1) {
            setShowWhenLocked(true)
            setTurnScreenOn(true)
        } else {
            @Suppress("DEPRECATION")
            window.addFlags(
                WindowManager.LayoutParams.FLAG_SHOW_WHEN_LOCKED or
                    WindowManager.LayoutParams.FLAG_TURN_SCREEN_ON
            )
        }
        binding = ActivityAlarmBinding.inflate(layoutInflater)
        setContentView(binding.root)
        readIntent(intent)
        binding.action.setOnClickListener { confirm() }
        binding.allowFullscreen.setOnClickListener {
            AlarmController.permissionIntent(this)?.let { startActivity(it) }
        }
        render()
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        setIntent(intent)
        readIntent(intent)
        render()
    }

    override fun onStart() {
        super.onStart()
        AlarmController.visible = true
        val filter = IntentFilter(WebSocketService.NEW_MESSAGE_BROADCAST)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            registerReceiver(receiver, filter, RECEIVER_NOT_EXPORTED)
        } else {
            @SuppressLint("UnspecifiedRegisterReceiverFlag")
            registerReceiver(receiver, filter)
        }
        clock.post(tick)
    }

    override fun onStop() {
        AlarmController.visible = false
        clock.removeCallbacks(tick)
        unregisterReceiver(receiver)
        super.onStop()
    }

    private fun readIntent(intent: Intent?) {
        val raw = intent?.getStringExtra(EXTRA_MESSAGE) ?: return
        message = Utils.JSON.fromJson(raw, Message::class.java)
        appName = intent.getStringExtra(EXTRA_APP_NAME) ?: appName
    }

    private fun confirm() {
        val current = message ?: return
        val incident = current.incident ?: return
        val next = nextStage(incident) ?: return
        binding.action.isEnabled = false
        val api = ClientFactory.clientToken(Settings(this)).createService(IncidentApi::class.java)
        api.advance(incident.id, AdvanceStage(next)).enqueue(
            Callback.callInUI(
                this,
                onSuccess = Callback.SuccessBody { updated ->
                    current.incident = updated
                    val raw = Utils.JSON.toJson(current)
                    AlarmController.onIncident(this, raw, appName)
                    message = current
                    binding.action.isEnabled = true
                    render()
                },
                onError = {
                    binding.action.isEnabled = true
                    Utils.showSnackBar(this, getString(R.string.oops))
                }
            )
        )
    }

    private fun render() {
        val current = message ?: return
        val incident = current.incident ?: return
        binding.clock.text = LocalTime.now().format(DateTimeFormatter.ofPattern("HH:mm"))
        binding.appName.text = appName
        binding.title.text = incident.title ?: current.title
        binding.body.text = incident.message ?: current.message
        paint(incident)
        binding.allowFullscreen.visibility = if (AlarmController.canUseFullScreen(this)) {
            android.view.View.GONE
        } else {
            android.view.View.VISIBLE
        }
    }

    private fun paint(incident: Incident) {
        val state = incident.state ?: ""
        val white = Color.WHITE
        when (state) {
            "waiting" -> {
                binding.root.setBackgroundColor(ContextCompat.getColor(this, R.color.colorPrimary))
                paintText(white)
                binding.done.visibility = android.view.View.VISIBLE
                val doneIndex = incident.currentStage ?: -1
                val doneLabel = incident.stages?.getOrNull(doneIndex)?.label ?: ""
                binding.done.text = doneLabel
                binding.status.text = deadlineLine(incident)
                showAction(incident)
                window.addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
            }
            "closed" -> {
                binding.root.setBackgroundColor(
                    MaterialColors.getColor(this, com.google.android.material.R.attr.colorSurface, Color.WHITE)
                )
                paintText(
                    MaterialColors.getColor(this, com.google.android.material.R.attr.colorOnSurface, Color.BLACK)
                )
                binding.done.visibility = android.view.View.GONE
                binding.action.visibility = android.view.View.GONE
                val closedAt = message?.date ?: OffsetDateTime.now()
                binding.status.text = getString(
                    R.string.alarm_closed,
                    closedAt.format(DateTimeFormatter.ofPattern("HH:mm"))
                )
                window.clearFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
            }
            else -> {
                binding.root.setBackgroundColor(ContextCompat.getColor(this, R.color.alarmRinging))
                paintText(white)
                binding.done.visibility = android.view.View.GONE
                binding.status.setText(R.string.alarm_ringing)
                showAction(incident)
                window.addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
            }
        }
    }

    private fun showAction(incident: Incident) {
        val next = nextStage(incident)
        val label = if (next == null) {
            null
        } else {
            incident.stages?.getOrNull(next)?.label
        }
        if (label == null) {
            binding.action.visibility = android.view.View.GONE
            return
        }
        binding.action.visibility = android.view.View.VISIBLE
        binding.action.text = label
        binding.action.backgroundTintList = android.content.res.ColorStateList.valueOf(Color.WHITE)
        binding.action.setTextColor(Color.BLACK)
    }

    private fun deadlineLine(incident: Incident): String {
        val deadline = incident.stageDeadlineAt ?: return getString(R.string.alarm_ringing)
        val remaining = Duration.between(OffsetDateTime.now(), deadline)
        val seconds = remaining.seconds.coerceAtLeast(0)
        val formatted = "%d:%02d".format(seconds / 60, seconds % 60)
        return getString(R.string.alarm_rings_again, formatted)
    }

    private fun paintText(color: Int) {
        binding.clock.setTextColor(color)
        binding.appName.setTextColor(color)
        binding.title.setTextColor(color)
        binding.body.setTextColor(color)
        binding.done.setTextColor(color)
        binding.status.setTextColor(color)
    }

    companion object {
        const val EXTRA_MESSAGE = "alarm-message"
        const val EXTRA_APP_NAME = "alarm-app-name"

        private fun nextStage(incident: Incident): Int? {
            val count = incident.stages?.size ?: return null
            val next = if (incident.state == "waiting") {
                (incident.currentStage ?: -1) + 1
            } else if (incident.state == "ringing") {
                0
            } else {
                return null
            }
            if (next < 0 || next >= count) {
                return null
            }
            return next
        }

        private fun shouldApply(current: Incident?, incoming: Incident): Boolean {
            if (current == null || current.id != incoming.id) {
                return true
            }
            if (current.state == "waiting" && incoming.kind == "open") {
                return false
            }
            if (current.state == "closed" && incoming.kind != "open" && incoming.kind != "reopen") {
                return false
            }
            return true
        }
    }
}
