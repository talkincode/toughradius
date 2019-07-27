package org.toughradius.config;

import freemarker.template.TemplateException;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.ViewResolver;
import org.springframework.web.servlet.config.annotation.*;
import org.springframework.web.servlet.view.freemarker.FreeMarkerConfigurer;
import org.springframework.web.servlet.view.freemarker.FreeMarkerViewResolver;

import java.io.IOException;

@Configuration
public class MvcConfigurer extends WebMvcConfigurerAdapter {

    @Autowired
    private PortalConfig portalConfig;

    @Autowired
    private AccessInterceptor accessInterceptor;


    @Override
    public void addViewControllers(ViewControllerRegistry registry) {
        registry.addViewController("/error").setViewName("/templates/global_error.html");
        super.addViewControllers(registry);
    }

    @Override
    public void configurePathMatch(PathMatchConfigurer configurer) {
        super.configurePathMatch(configurer);
        configurer.setUseSuffixPatternMatch(false);//当此参数设置为true的时候，那么/user.html，/user.aa，/user.*都能是正常访问的。
    }

    public void addInterceptors(InterceptorRegistry registry) {
        registry.addInterceptor(accessInterceptor).addPathPatterns("/api/v6/**");
        registry.addInterceptor(new SessionInterceptor()).addPathPatterns("/admin/**")
                .excludePathPatterns("/")
                .excludePathPatterns("/admin")
                .excludePathPatterns("/admin/verify-img.jpg")
                .excludePathPatterns("/admin/login")
                .excludePathPatterns("/admin/session")
                .excludePathPatterns("/admin/logout");
        super.addInterceptors(registry);
    }

    @Override
    public void addResourceHandlers(ResourceHandlerRegistry registry) {
        registry.addResourceHandler("/static/**").addResourceLocations("classpath:/static/");
        registry.addResourceHandler("/portal/**").addResourceLocations(portalConfig.getTemplateDir());
        super.addResourceHandlers(registry);
    }

    @Bean
    public ViewResolver viewResolver() {
        FreeMarkerViewResolver resolver = new FreeMarkerViewResolver();
        resolver.setCache(true);
        resolver.setPrefix("");
        resolver.setSuffix(".html");
        resolver.setContentType("text/html; charset=UTF-8");
        return resolver;
    }

    @Bean
    public FreeMarkerConfigurer freemarkerConfig() throws IOException, TemplateException {
        FreeMarkerConfigurer configurer = new FreeMarkerConfigurer();
        configurer.setTemplateLoaderPaths(portalConfig.getTemplateDir(),"classpath:/templates/");
        configurer.setDefaultEncoding("UTF-8");
        return configurer;
    }


}
