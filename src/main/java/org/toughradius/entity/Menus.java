package org.toughradius.entity;

import java.util.List;

public class Menus {

    private String id;

    private  String icon;

    private String value;

    private List<Menus> data;

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getIcon() {
        return icon;
    }

    public void setIcon(String icon) {
        this.icon = icon;
    }

    public String getValue() {
        return value;
    }

    public void setValue(String value) {
        this.value = value;
    }

    public List<Menus> getData() {
        return data;
    }

    public void setData(List<Menus> data) {
        this.data = data;
    }
}
