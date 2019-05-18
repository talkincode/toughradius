package org.toughradius.common;

import java.util.List;

public class PageResult<T> {
    private long pos;
    private long total_count;
    private List<T> data;

    public PageResult(long pos, long total_count, List<T> data) {
        this.pos = pos;
        this.total_count = total_count;
        this.data = data;
    }

    public long getPos() {
        return pos;
    }

    public void setPos(int pos) {
        this.pos = pos;
    }

    public long getTotal_count() {
        return total_count;
    }

    public void setTotal_count(int total_count) {
        this.total_count = total_count;
    }

    public List<T> getData() {
        return data;
    }

    public void setData(List<T> data) {
        this.data = data;
    }
}


