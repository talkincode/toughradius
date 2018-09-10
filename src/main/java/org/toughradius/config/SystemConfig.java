package org.toughradius.config;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Configuration;

@Configuration
@ConfigurationProperties(prefix = "ts.system")
public class SystemConfig {

    private String version;
    private String superRpcUrl;
    private String superRpcUser;
    private String superRpcPwd;
    private String datadir;
    private boolean trace;

    public String getVersion() {
        return version;
    }

    public void setVersion(String version) {
        this.version = version;
    }

    public String getSuperRpcUrl() {
        return superRpcUrl;
    }

    public void setSuperRpcUrl(String superRpcUrl) {
        this.superRpcUrl = superRpcUrl;
    }

    public String getSuperRpcUser() {
        return superRpcUser;
    }

    public void setSuperRpcUser(String superRpcUser) {
        this.superRpcUser = superRpcUser;
    }

    public String getSuperRpcPwd() {
        return superRpcPwd;
    }

    public void setSuperRpcPwd(String superRpcPwd) {
        this.superRpcPwd = superRpcPwd;
    }

    public String getDatadir() {
        return datadir;
    }

    public void setDatadir(String datadir) {
        this.datadir = datadir;
    }

    public boolean isTrace() {
        return trace;
    }

    public void setTrace(boolean trace) {
        this.trace = trace;
    }
}
