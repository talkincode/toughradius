package org.toughradius.common;

import java.util.List;

public class PageResult<T> {
    private int pos;
    private int total_count;
    private List<T> data;

    public PageResult(int pos, int total_count, List<T> data) {
        this.pos = pos;
        this.total_count = total_count;
        this.data = data;
    }

    public int getPos() {
        return pos;
    }

    public void setPos(int pos) {
        this.pos = pos;
    }

    public int getTotal_count() {
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


