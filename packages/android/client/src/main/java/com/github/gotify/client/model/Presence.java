package com.github.gotify.client.model;

import com.google.gson.annotations.SerializedName;
import java.time.OffsetDateTime;
import java.util.Objects;

/** Latest presence report of a user. */
public class Presence {
    @SerializedName("userId")
    private Long userId;

    @SerializedName("state")
    private String state;

    @SerializedName("source")
    private String source;

    @SerializedName("detail")
    private String detail;

    @SerializedName("reportedAt")
    private OffsetDateTime reportedAt;

    @SerializedName("validUntil")
    private OffsetDateTime validUntil;

    public Long getUserId() {
        return userId;
    }

    public String getState() {
        return state;
    }

    public String getSource() {
        return source;
    }

    public String getDetail() {
        return detail;
    }

    public OffsetDateTime getReportedAt() {
        return reportedAt;
    }

    public OffsetDateTime getValidUntil() {
        return validUntil;
    }

    @Override
    public boolean equals(Object other) {
        if (this == other) {
            return true;
        }
        if (other == null || getClass() != other.getClass()) {
            return false;
        }
        Presence presence = (Presence) other;
        return Objects.equals(userId, presence.userId)
                && Objects.equals(state, presence.state)
                && Objects.equals(source, presence.source)
                && Objects.equals(detail, presence.detail)
                && Objects.equals(reportedAt, presence.reportedAt)
                && Objects.equals(validUntil, presence.validUntil);
    }

    @Override
    public int hashCode() {
        return Objects.hash(userId, state, source, detail, reportedAt, validUntil);
    }

    @Override
    public String toString() {
        return "Presence{userId=" + userId + ", state='" + state + "', source='" + source + "'}";
    }
}
