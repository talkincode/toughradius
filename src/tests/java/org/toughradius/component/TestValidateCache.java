package org.toughradius.component;

import org.toughradius.common.ValidateCache;
import org.junit.Test;

public class TestValidateCache {

    @Test
    public void testValidate(){
        ValidateCache vc = new ValidateCache(2000, 5);
        vc.incr("tests");
        vc.incr("tests");
        vc.incr("tests");
        assert !vc.isOver("tests");
        vc.incr("tests");
        vc.incr("tests");
        vc.incr("tests");
        System.out.println(vc.errors("tests"));
        assert vc.isOver("tests");
        try {
            Thread.sleep(2001);
        } catch (InterruptedException e) {
            e.printStackTrace();
        }
        assert !vc.isOver("tests");
    }
}
