package org.toughradius;

import org.junit.Test;

import java.util.concurrent.atomic.AtomicInteger;

public class AtomIntTest {

    @Test
    public void TestIncr(){
        AtomicInteger a = new AtomicInteger(0);
        a.addAndGet(10);
        a.addAndGet(10);
        assert a.intValue() == 20;
    }
}
