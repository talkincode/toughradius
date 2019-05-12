package org.toughradius.component;

import org.toughradius.entity.Subscribe;
import org.toughradius.entity.SubscribeQuery;
import org.toughradius.mapper.SubscribeMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.math.BigInteger;
import java.util.List;

@Service
public class SubscribeService {

    @Autowired
    private SubscribeMapper subscribeMapper;

    public Subscribe findSubscribe(String username){
        return subscribeMapper.findSubscribe(username);
    }

    public void updateFlowAmountByUsername(String username, BigInteger flowAmount){
        subscribeMapper.updateFlowAmountByUsername(username,flowAmount);
    }

    public int updateMacAddr(String username, String macAddr){
        return subscribeMapper.updateMacAddr(username,macAddr);
    }

    public int updateInValn(String username, Integer inValn){
        return subscribeMapper.updateInValn(username,inValn);
    }

    public int updateOutValn(String username, Integer outValn){
        return subscribeMapper.updateOutValn(username,outValn);
    }

    public List<Subscribe> findLastUpdateUser(String lastUpdate) {
        return subscribeMapper.findLastUpdateUser(lastUpdate);
    }

    public int startOnline(Long userid){
        return subscribeMapper.updateOnlineStatus(userid,1);
    }

    public int stopOnline(Long userid){
        return subscribeMapper.updateOnlineStatus(userid,0);
    }

    public List<Subscribe> queryForList(SubscribeQuery subscribe){
        return subscribeMapper.queryForList(subscribe);
    }
    public Integer queryTotal(SubscribeQuery subscribe){
        return subscribeMapper.queryTotal(subscribe);
    }

    public void insertSubscribe(Subscribe subscribe){
        subscribeMapper.insertSubscribe(subscribe);
    }

    public void updateSubscribe(Subscribe subscribe){
        subscribeMapper.updateSubscribe(subscribe);
    }

    public void updatePassword(Long id, String password){
        subscribeMapper.updatePassword(id, password);
    }

    public Subscribe findById(Long id){
        return subscribeMapper.findById(id);
    }
    public void deleteById(Long id){
        subscribeMapper.deleteById(id);
    }

}
