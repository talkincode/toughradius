package org.toughradius.config;

import com.google.gson.Gson;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.boot.web.servlet.MultipartConfigFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.scheduling.concurrent.ThreadPoolTaskExecutor;

import javax.servlet.MultipartConfigElement;

@Configuration
@ConfigurationProperties(prefix = "application")
public class ApplicationConfig {

    private String version;
    private String apikey;
    private String apisecret;

    @Bean
    public MultipartConfigElement multipartConfigElement() {
        MultipartConfigFactory factory = new MultipartConfigFactory();
        factory.setMaxFileSize("1024000KB");
        factory.setMaxRequestSize("102400000KB");
        return factory.createMultipartConfig();
    }

    @Bean
    public ThreadPoolTaskExecutor systaskExecutor(){
        ThreadPoolTaskExecutor taskExecutor = new ThreadPoolTaskExecutor();
        taskExecutor.setCorePoolSize(32);
        taskExecutor.setMaxPoolSize(512);
        taskExecutor.setQueueCapacity(100000);
        taskExecutor.setKeepAliveSeconds(60);
        taskExecutor.setThreadNamePrefix("TASK_EXECUTOR");
        return taskExecutor;
    }

    @Bean
    public Gson gson(){
        return new Gson();
    }

    public String getApikey() {
        return apikey;
    }

    public void setApikey(String apikey) {
        this.apikey = apikey;
    }

    public String getApisecret() {
        return apisecret;
    }

    public void setApisecret(String apisecret) {
        this.apisecret = apisecret;
    }

    public String getVersion() {
        return version;
    }

    public void setVersion(String version) {
        this.version = version;
    }
}
