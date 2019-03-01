package org.toughradius;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cache.annotation.EnableCaching;
import org.springframework.context.annotation.Configuration;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
//@EnableTransactionManagement()
@EnableScheduling
@Configuration
@EnableCaching
public class RadiusdApplication {

    public static void main(String[] args) throws Exception {
        SpringApplication.run(RadiusdApplication.class, args);
    }
}