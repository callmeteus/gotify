package com.github.gotify.client.model;

import com.google.gson.annotations.SerializedName;
import java.util.Objects;

/** Opt-in frame sent on /stream?presence=1. Message frames have no {@code type} field. */
public class PresenceFrame {
    @SerializedName("type")
    private String type;

    @SerializedName("data")
    private Presence data;

    public String getType() {
        return type;
    }

    public Presence getData() {
        return data;
    }

    public boolean isPresence() {
        return "presence".equals(type) && data != null;
    }

    @Override
    public boolean equals(Object other) {
        if (this == other) {
            return true;
        }
        if (other == null || getClass() != other.getClass()) {
            return false;
        }
        PresenceFrame frame = (PresenceFrame) other;
        return Objects.equals(type, frame.type) && Objects.equals(data, frame.data);
    }

    @Override
    public int hashCode() {
        return Objects.hash(type, data);
    }
}
