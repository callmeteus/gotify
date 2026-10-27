package com.github.gotify.client.model;

import com.google.gson.annotations.SerializedName;
import java.util.Objects;

/** One button on the alarm screen. */
public class AlarmStage {
    public static final String SERIALIZED_NAME_LABEL = "label";

    @SerializedName(SERIALIZED_NAME_LABEL)
    private String label;

    public static final String SERIALIZED_NAME_ADVANCE_WITHIN_SECONDS = "advanceWithinSeconds";

    @SerializedName(SERIALIZED_NAME_ADVANCE_WITHIN_SECONDS)
    private Integer advanceWithinSeconds;

    public String getLabel() {
        return label;
    }

    public void setLabel(String label) {
        this.label = label;
    }

    public Integer getAdvanceWithinSeconds() {
        return advanceWithinSeconds;
    }

    public void setAdvanceWithinSeconds(Integer advanceWithinSeconds) {
        this.advanceWithinSeconds = advanceWithinSeconds;
    }

    @Override
    public boolean equals(Object other) {
        if (this == other) {
            return true;
        }
        if (other == null || getClass() != other.getClass()) {
            return false;
        }
        AlarmStage stage = (AlarmStage) other;
        return Objects.equals(label, stage.label)
                && Objects.equals(advanceWithinSeconds, stage.advanceWithinSeconds);
    }

    @Override
    public int hashCode() {
        return Objects.hash(label, advanceWithinSeconds);
    }
}
