package org.toughradius.component;

import org.toughradius.entity.Subscribe;
import org.toughradius.entity.SubscribeBill;
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

    public SubscribeBill fetchSubscribeBill(String username) {
        return subscribeMapper.fetchSubscribeBill(username);
    }

    public int startOnline(String username){
        return subscribeMapper.updateOnlineStatus(username,1);
    }

    public int stopOnline(String username){
        return subscribeMapper.updateOnlineStatus(username,0);
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

    public Subscribe findById(Integer id){
        return subscribeMapper.findById(id);
    }
    public void deleteById(Integer id){
        subscribeMapper.deleteById(id);
    }

}
