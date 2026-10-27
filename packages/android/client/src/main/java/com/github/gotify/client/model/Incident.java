package com.github.gotify.client.model;

import com.google.gson.annotations.SerializedName;
import java.time.OffsetDateTime;
import java.util.List;
import java.util.Objects;

/** Alarm incident attached to a message or returned by the incident routes. */
public class Incident {
    @SerializedName("id")
    private Long id;

    @SerializedName("appid")
    private Long appid;

    @SerializedName("state")
    private String state;

    @SerializedName("kind")
    private String kind;

    @SerializedName("currentStage")
    private Integer currentStage;

    @SerializedName("stages")
    private List<AlarmStage> stages;

    @SerializedName("stageDeadlineAt")
    private OffsetDateTime stageDeadlineAt;

    @SerializedName("ringUntil")
    private OffsetDateTime ringUntil;

    @SerializedName("repeatIntervalSeconds")
    private Integer repeatIntervalSeconds;

    @SerializedName("title")
    private String title;

    @SerializedName("message")
    private String message;

    public Long getId() {
        return id;
    }

    public Long getAppid() {
        return appid;
    }

    public String getState() {
        return state;
    }

    public String getKind() {
        return kind;
    }

    public Integer getCurrentStage() {
        return currentStage;
    }

    public List<AlarmStage> getStages() {
        return stages;
    }

    public OffsetDateTime getStageDeadlineAt() {
        return stageDeadlineAt;
    }

    public OffsetDateTime getRingUntil() {
        return ringUntil;
    }

    public Integer getRepeatIntervalSeconds() {
        return repeatIntervalSeconds;
    }

    public String getTitle() {
        return title;
    }

    public String getMessage() {
        return message;
    }

    @Override
    public boolean equals(Object other) {
        if (this == other) {
            return true;
        }
        if (other == null || getClass() != other.getClass()) {
            return false;
        }
        Incident incident = (Incident) other;
        return Objects.equals(id, incident.id)
                && Objects.equals(appid, incident.appid)
                && Objects.equals(state, incident.state)
                && Objects.equals(kind, incident.kind)
                && Objects.equals(currentStage, incident.currentStage)
                && Objects.equals(stages, incident.stages)
                && Objects.equals(stageDeadlineAt, incident.stageDeadlineAt)
                && Objects.equals(ringUntil, incident.ringUntil)
                && Objects.equals(repeatIntervalSeconds, incident.repeatIntervalSeconds)
                && Objects.equals(title, incident.title)
                && Objects.equals(message, incident.message);
    }

    @Override
    public int hashCode() {
        return Objects.hash(
                id,
                appid,
                state,
                kind,
                currentStage,
                stages,
                stageDeadlineAt,
                ringUntil,
                repeatIntervalSeconds,
                title,
                message);
    }
}
