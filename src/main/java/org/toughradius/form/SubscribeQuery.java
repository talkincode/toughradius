package org.toughradius.form;

import org.toughradius.entity.Subscribe;

public class SubscribeQuery extends Subscribe {

    private String keyword;

    public String getKeyword() {
        return keyword;
    }

    public void setKeyword(String keyword) {
        this.keyword = keyword;
    }
}
