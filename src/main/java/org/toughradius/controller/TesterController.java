package org.toughradius.controller;

import org.toughradius.component.RadiusTester;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class TesterController {

    @Autowired
    private RadiusTester radiusTester;


    @GetMapping("/admin/radius/auth/test")
    public String AuthTestHandler(String username,String papchap){
        return radiusTester.sendAuth(username,papchap).replaceAll("\n","<br>");
    }


    @GetMapping("/admin/radius/acct/test")
    public String AcctTestHandler(String username,int type){
        return radiusTester.sendAcct(username,type).replaceAll("\n","<br>");
    }

}
