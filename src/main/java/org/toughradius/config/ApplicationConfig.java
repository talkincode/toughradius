package org.toughradius.config;

import org.springframework.boot.web.servlet.MultipartConfigFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.scheduling.concurrent.ThreadPoolTaskExecutor;

import javax.servlet.MultipartConfigElement;

@Configuration
public class ApplicationConfig {


    @Bean
    public MultipartConfigElement multipartConfigElement() {
        MultipartConfigFactory factory = new MultipartConfigFactory();
        factory.setMaxFileSize("1024000KB");
        factory.setMaxRequestSize("102400000KB");
        return factory.createMultipartConfig();
    }

    @Bean
    public ThreadPoolTaskExecutor taskExecutor(){
        ThreadPoolTaskExecutor taskExecutor = new ThreadPoolTaskExecutor();
        taskExecutor.setCorePoolSize(32);
        taskExecutor.setMaxPoolSize(512);
        taskExecutor.setQueueCapacity(100000);
        taskExecutor.setKeepAliveSeconds(60);
        taskExecutor.setThreadNamePrefix("TASK_EXECUTOR");
        return taskExecutor;
    }
}
