package org.toughradius.config;

import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.context.junit4.SpringRunner;

@RunWith(SpringRunner.class)
@SpringBootTest
@ActiveProfiles("dev")
public class RedisCacheTest {


    @Autowired
    private RedisTemplate<Object, Object> redisTemplate;

    @Autowired
    private StringRedisTemplate stringRedisTemplate;


    @Test
    public void testString(){
        stringRedisTemplate.opsForValue().set(RedisConfig.testKey("strkey"),"testvalue");

        String val = stringRedisTemplate.opsForValue().get(RedisConfig.testKey("strkey"));
        assert val.equals("testvalue");
    }

    @Test
    public void testList(){
        stringRedisTemplate.opsForList().leftPush(RedisConfig.testKey("listkey"),"testvalue");
        String val = stringRedisTemplate.opsForList().leftPop(RedisConfig.testKey("listkey"));
        assert val.equals("testvalue");
    }

    @Test
    public void testCutList(){
        for (int i =0;i<20;i++){
            stringRedisTemplate.opsForList().rightPush(RedisConfig.testKey("listkey"),"testvalue"+i);
        }
        stringRedisTemplate.opsForList().trim(RedisConfig.testKey("listkey"),0,9);
        assert stringRedisTemplate.opsForList().size(RedisConfig.testKey("listkey")) == 10;
    }

//    @Test
//    public void testObjList(){
//        LineChartModel lm = new LineChartModel(100L,100L);
//        LineChartModel lm2 = new LineChartModel(100L,100L);
//        redisTemplate.opsForList().leftPush(RedisConfig.testKey("olistkey"),lm);
//        redisTemplate.opsForList().leftPush(RedisConfig.testKey("olistkey"),lm2);
//        LineChartModel val = (LineChartModel) redisTemplate.opsForList().leftPop(RedisConfig.testKey("olistkey"));
//        assert val.getValue() == lm.getValue();
//    }



}
