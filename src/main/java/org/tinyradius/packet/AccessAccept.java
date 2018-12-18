package org.tinyradius.packet;

import java.util.ArrayList;

public class AccessAccept extends RadiusPacket{

    private long preSessionTimeout;
    private int preInterim;

    public AccessAccept(final int identifier) {
        super(ACCESS_ACCEPT, identifier, new ArrayList());
    }

    public long getPreSessionTimeout() {
        return preSessionTimeout;
    }

    public void setPreSessionTimeout(long preSessionTimeout) {
        this.preSessionTimeout = preSessionTimeout;
    }

    public int getPreInterim() {
        return preInterim;
    }

    public void setPreInterim(int preInterim) {
        this.preInterim = preInterim;
    }
}
