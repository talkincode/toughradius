package org.toughradius.component;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.toughradius.entity.User;
import org.toughradius.mapper.UserMapper;

import java.math.BigInteger;
import java.util.List;

@Service
public class UserService {

    @Autowired
    private UserMapper userMapper;

    public User findUser(String username){
        return userMapper.findUser(username);
    }

    public void updateFlowAmountByUsername(String username, Long flowAmount){
        userMapper.updateFlowAmountByUsername(username,flowAmount);
    }

    public int updateMacAddr(String username, String macAddr){
        return userMapper.updateMacAddr(username,macAddr);
    }

    public int updateInValn(String username, Integer inValn){
        return userMapper.updateInValn(username,inValn);
    }

    public int updateOutValn(String username, Integer outValn){
        return userMapper.updateOutValn(username,outValn);
    }

    public List<User> findLastUpdateUser(String lastUpdate) {
        return userMapper.findLastUpdateUser(lastUpdate);
    }

    public BigInteger getFlowAmount(String username) {
        return userMapper.getFlowAmount(username);
    }
}
