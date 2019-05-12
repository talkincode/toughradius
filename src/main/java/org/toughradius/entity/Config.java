package org.toughradius.entity;

public class Config {
    private Long id;

    private String type;

    private String name;

    private String value;

    private String remark;

    public Config() {
    }

    public Config(String type, String name, String value) {
        this.type = type;
        this.name = name;
        this.value = value;
    }

    public Config(String type, String name, String value, String remark) {
        this.type = type;
        this.name = name;
        this.value = value;
        this.remark = remark;
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getType() {
        return type;
    }

    public void setType(String type) {
        this.type = type == null ? null : type.trim();
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name == null ? null : name.trim();
    }

    public String getValue() {
        return value;
    }

    public void setValue(String value) {
        this.value = value == null ? null : value.trim();
    }

    public String getRemark() {
        return remark;
    }

    public void setRemark(String remark) {
        this.remark = remark == null ? null : remark.trim();
    }
}