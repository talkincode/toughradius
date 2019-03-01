package org.toughradius;

import org.junit.Test;

public class UsernameTest {

    @Test
    public void testDomainUser(){
        String duser = "username@domain";
        assert duser.substring(0,duser.indexOf("@")).equalsIgnoreCase("username");
    }
}
