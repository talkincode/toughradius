package org.toughradius.config;

import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.*;

@Configuration
public class MvcConfigurer extends WebMvcConfigurerAdapter {


    @Override
    public void addViewControllers(ViewControllerRegistry registry) {
        registry.addViewController("/accessError").setViewName("/static/accesseErrorPage.html");
        super.addViewControllers(registry);
    }

    @Override
    public void configurePathMatch(PathMatchConfigurer configurer) {
        super.configurePathMatch(configurer);
        configurer.setUseSuffixPatternMatch(false);//当此参数设置为true的时候，那么/user.html，/user.aa，/user.*都能是正常访问的。
    }

    public void addInterceptors(InterceptorRegistry registry) {
        registry.addInterceptor(new AccessInterceptor()).addPathPatterns("/api/**");
        super.addInterceptors(registry);
    }


    @Override
    public void addResourceHandlers(ResourceHandlerRegistry registry) {
        registry.addResourceHandler("/static/**").addResourceLocations("classpath:/static/");
        super.addResourceHandlers(registry);
    }
}
