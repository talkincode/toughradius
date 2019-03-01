package org.toughradius.component;


import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.junit4.SpringRunner;

@RunWith(SpringRunner.class)
@SpringBootTest
public class SysLoggerTest {

    @Autowired
    private Syslogger logger;

    @Test
    public void testTrace(){
        String message = "tests message";
        logger.trace(message,Syslogger.RADIUSD);
        logger.info(message,Syslogger.RADIUSD);
    }
}
