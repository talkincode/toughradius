package org.toughradius.component;

import com.google.gson.Gson;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.junit4.SpringRunner;
import org.toughradius.component.RadiusStat;

@RunWith(SpringRunner.class)
@SpringBootTest
public class RadiusStatTest {

    @Autowired
    private RadiusStat stat;

    @Autowired
    private Gson gson;

    @Test
    public  void TestStat(){
        System.out.println(gson.toJson(stat.getData()));
    }
}
