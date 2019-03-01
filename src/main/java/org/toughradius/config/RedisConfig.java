package org.toughradius.config;

import com.fasterxml.jackson.annotation.JsonAutoDetect;
import com.fasterxml.jackson.annotation.PropertyAccessor;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.commons.pool2.impl.GenericObjectPoolConfig;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.autoconfigure.condition.ConditionalOnMissingBean;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.core.env.Environment;
import org.springframework.data.redis.connection.RedisPassword;
import org.springframework.data.redis.connection.RedisStandaloneConfiguration;
import org.springframework.data.redis.connection.jedis.JedisConnectionFactory;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.data.redis.serializer.Jackson2JsonRedisSerializer;
import org.springframework.data.redis.serializer.StringRedisSerializer;

import java.net.UnknownHostException;


@Configuration
public class RedisConfig {

    @Autowired
    private Environment env;

    public final  static String PREFIX = "org.toughradius:key:";

    @Bean
    public JedisConnectionFactory jedisConnectionFactory(){
        JedisConnectionFactory factory = new JedisConnectionFactory();
        RedisStandaloneConfiguration cfg = factory.getStandaloneConfiguration();
        cfg.setHostName(env.getRequiredProperty("spring.redis.host"));
        cfg.setPort(Integer.parseInt(env.getRequiredProperty("spring.redis.port")));
        cfg.setPassword(RedisPassword.of(env.getProperty("spring.redis.password")));
        cfg.setDatabase(Integer.parseInt(env.getRequiredProperty("spring.redis.database")));
        GenericObjectPoolConfig poolcfg = factory.getPoolConfig();
        poolcfg.setMaxTotal(Integer.parseInt(env.getRequiredProperty("spring.redis.jedis.pool.max-active")));
        poolcfg.setMinIdle(Integer.parseInt(env.getRequiredProperty("spring.redis.jedis.pool.min-idle")));
        poolcfg.setMaxIdle(Integer.parseInt(env.getRequiredProperty("spring.redis.jedis.pool.max-idle")));
        return factory;
    }

    @Bean
    @ConditionalOnMissingBean(name="redisTemplate")
    public RedisTemplate<Object, Object> redisTemplate(JedisConnectionFactory jedisConnectionFactory)throws UnknownHostException {
        RedisTemplate<Object, Object> template = new RedisTemplate<Object,Object>();
        template.setConnectionFactory(jedisConnectionFactory);
        Jackson2JsonRedisSerializer jackson2JsonRedisSerializer = new Jackson2JsonRedisSerializer(Object.class);
        ObjectMapper objectMapper = new ObjectMapper();
        objectMapper.setVisibility(PropertyAccessor.ALL, JsonAutoDetect.Visibility.ANY);
        objectMapper.enableDefaultTyping(ObjectMapper.DefaultTyping.NON_FINAL);
        jackson2JsonRedisSerializer.setObjectMapper(objectMapper);
        // 设置value的序列化规则和 key的序列化规则
        template.setKeySerializer(new StringRedisSerializer());
        template.setValueSerializer(jackson2JsonRedisSerializer);
        template.afterPropertiesSet();
        return template;
    }

    @Bean
    @ConditionalOnMissingBean(StringRedisTemplate.class)
    public StringRedisTemplate stringRedisTemplate(JedisConnectionFactory jedisConnectionFactory)throws UnknownHostException{
        StringRedisTemplate template = new StringRedisTemplate();
        template.setConnectionFactory(jedisConnectionFactory);
        Jackson2JsonRedisSerializer jackson2JsonRedisSerializer = new Jackson2JsonRedisSerializer(Object.class);
        ObjectMapper om = new ObjectMapper();
        om.setVisibility(PropertyAccessor.ALL, JsonAutoDetect.Visibility.ANY);
        om.enableDefaultTyping(ObjectMapper.DefaultTyping.NON_FINAL);
        jackson2JsonRedisSerializer.setObjectMapper(om);
        template.setValueSerializer(jackson2JsonRedisSerializer);
        template.setHashValueSerializer(jackson2JsonRedisSerializer);
        template.afterPropertiesSet();
        return template;
    }


    public static String newKey(String keystr){
        return String.format("%s%s",PREFIX, keystr);
    }
    public static String testKey(String keystr){
        return String.format("%stest:%s",PREFIX, keystr);
    }

}
