package org.toughradius.component;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.concurrent.ThreadPoolTaskExecutor;
import org.springframework.stereotype.Service;
import org.tinyradius.packet.AccountingRequest;
import org.tinyradius.packet.CoaRequest;
import org.tinyradius.packet.RadiusPacket;
import org.tinyradius.util.RadiusClient;
import org.tinyradius.util.RadiusException;
import org.toughradius.common.DateTimeUtil;
import org.toughradius.common.ValidateUtil;
import org.toughradius.entity.Nas;
import org.toughradius.entity.Online;
import org.toughradius.entity.PageResult;

import java.io.IOException;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class OnlineCache {

    private final static  ConcurrentHashMap<String,Online> cacheData = new ConcurrentHashMap<String,Online>();
    private final static Log logger = LogFactory.getLog(OnlineCache.class);

    @Autowired
    private NasService nasService;

    @Autowired
    private ThreadPoolTaskExecutor taskExecutor;


    public ConcurrentHashMap<String,Online> getCacheData(){
        return cacheData;
    }

    public int size()
    {
        return cacheData.size();
    }

    public List<Online> getOnlineList()
    {
        return new ArrayList<Online>(Collections.unmodifiableCollection(cacheData.values()));
    }

    public List<Online> list(String groupId)
    {
        List<Online> onlineList = new ArrayList<Online>();
        for (Online online : cacheData.values()) {
            if (online != null && online.getGroupId().equals(groupId))
                onlineList.add(online);
        }
        return onlineList;
    }

    public List<Online> queryNoAmountOnline( )
    {
        List<Online> onlineList = new ArrayList<Online>();
        for (Online online : cacheData.values()) {
            if (online.getUnLockFlag() == Online.AMOUNT_NOT_ENOUGH)
                onlineList.add(online);
        }
        return onlineList;
    }

    public boolean isExist(String sessionId)
    {
        return cacheData.containsKey(sessionId);
    }

    public Online getOnline(String sessionId)
    {
        return cacheData.get(sessionId);
    }

    public boolean isOnline(String userName)
    {
        boolean isOnline = false;
        for (Online online : cacheData.values()) {
            if (userName.equalsIgnoreCase(online.getUsername())) {
                isOnline = true;
                break;
            }
        }
        return isOnline;
    }

    public Online getFirstOnline(String userName)
    {
        for (Online online : cacheData.values()) {
            if (userName.equalsIgnoreCase(online.getUsername())) {
                return online;
            }
        }
        return null;
    }


    public Online getLastOnline(String userName)
    {
        Online online =null;
        for (Online _online : cacheData.values()) {
            if (userName.equalsIgnoreCase(_online.getUsername())) {
                online = _online;
                break;
            }
        }
        return online;
    }


    /**
     * 异步批量下线
     * @param ids
     */
    public void unlockOnlines(String ids)
    {
        for(String sessionid : ids.split(",")){
            Online online = getOnline(sessionid);
            if(online==null){
                logger.error("发送下线失败,无在线信息");
                continue;
            }
            taskExecutor.execute(()->{
                try {
                    Nas bras = nasService.findNas(online.getNasAddr(),online.getNasId());
                    RadiusClient cli = new RadiusClient(online.getNasPaddr(),bras.getSecret());
                    CoaRequest dmreq = new CoaRequest(RadiusPacket.DISCONNECT_REQUEST);
                    dmreq.addAttribute("User-Name",online.getUsername());
                    dmreq.addAttribute("Acct-Session-Id",online.getAcctSessionId());
                    if(ValidateUtil.isNotEmpty(online.getNasAddr())&&!online.getNasAddr().equals("0.0.0.0"))
                        dmreq.addAttribute("NAS-IP-Address",online.getNasAddr());
                    logger.info(String.format("发送下线请求 %s", dmreq.toString()));
                    RadiusPacket dmrep = cli.communicate(dmreq,bras.getCoaPort());
                    logger.info(String.format("接收到下线响应 %s", dmrep.toString()));
                } catch (ServiceException | IOException | RadiusException e) {
                    logger.error("发送下线失败",e);
                    removeOnline(sessionid);
                }
            });
        }
    }

    /**
     * 单个下线
     * @param sessionId
     */
    public boolean unlockOnline(String sessionId)
    {
        Online online = getOnline(sessionId);
        try {
            if(online==null){
                logger.error("发送下线失败,无在线信息");
                return false;
            }
            Nas bras = nasService.findNas(online.getNasId(), online.getNasAddr());
            RadiusClient cli = new RadiusClient(online.getNasPaddr(),bras.getSecret());
            CoaRequest dmreq = new CoaRequest(RadiusPacket.DISCONNECT_REQUEST);
            dmreq.addAttribute("User-Name",online.getUsername());
            dmreq.addAttribute("Acct-Session-Id",online.getAcctSessionId());
            dmreq.addAttribute("NAS-IP-Address",online.getNasAddr());
            logger.info(String.format("发送下线请求 %s", dmreq.toString()));
            RadiusPacket dmrep = cli.communicate(dmreq,bras.getCoaPort());
            logger.info(String.format("接收到下线响应 %s", dmrep.toString()));
            return dmrep.getPacketType() == RadiusPacket.DISCONNECT_ACK;
        } catch (ServiceException | IOException | RadiusException e) {
            logger.error("发送下线失败",e);
            removeOnline(sessionId);
            return false;
        }
    }


    public List<Online> getOnlineByUserName(String userName)
    {
        List<Online> onlineList = new ArrayList<Online>();
        for (Online online : cacheData.values()) {
            if (userName.equalsIgnoreCase(online.getUsername())) {
                onlineList.add(online);
            }
        }
        return onlineList;
    }

    /** 用户上线 */
    public void putOnline(Online online)
    {
        synchronized (cacheData)
        {
            String key = online.getAcctSessionId();
            cacheData.put(key, online);
        }
    }

    /** 一个用户下线 */
    public Online removeOnline(String sessionId)
    {
        return (Online) cacheData.remove(sessionId);
    }

    /** 设置解锁标记 */
    public void setUnLock(String sessionId,int  flag)
    {
        Online online = (Online) cacheData.get(sessionId);
        if(online!=null){
            online.setUnLockFlag(flag);
        }
    }

    /** BAS所有在线用户下线 */
    public List<Online> removeAllOnline(String nasAddr, String nasId)
    {
        List<Online> onlineList = new ArrayList<Online>();
        synchronized (cacheData)
        {
            for (Iterator<Online> it = cacheData.values().iterator(); it.hasNext();)
            {
                Online online = it.next();
                if (online.getNasId().equalsIgnoreCase(nasId)|| online.getNasAddr().equalsIgnoreCase(nasAddr)||online.getNasPaddr().equalsIgnoreCase(nasAddr))
                {
                    onlineList.add(online);
                    it.remove();
                }
            }
        }
        return onlineList;
    }

    /** 查询上网帐号并发数 */
    public int getUserOnlineNum(String userName)
    {
        int onlineNum = 0;
        for (Online online : cacheData.values()) {
            if (userName.equalsIgnoreCase(online.getUsername()))
                onlineNum++;
        }
        return onlineNum;
    }

    /** 查询上网帐号并发数,要求MAC地址不相等 */
    public int getUserOnlineNum(String userName, String macAddr)
    {
        int onlineNum = 0;
        for (Online online : cacheData.values()) {
            if (userName.equalsIgnoreCase(online.getUsername()) && !macAddr.equalsIgnoreCase(online.getMacAddr()))
                onlineNum++;
        }
        return onlineNum;
    }

    /**
     * 间隔 interim_times 不更新的为过期
     * @param online
     * @param interim_times
     * @return
     */
    private  boolean isExpire(Online online, int interim_times)
    {
        String curTime = DateTimeUtil.getDateTimeString();
        String acctStart =  online.getAcctStartTime();
        int second = DateTimeUtil.compareSecond(curTime,acctStart);
        if (second > (online.getAcctSessionTime()+interim_times+120))
            return true;
        else
            return false;
    }

    public void clearOvertimeOnline( int interim_times)
    {
        for (Iterator<Online> it = cacheData.values().iterator(); it.hasNext();)
        {
            Online online = it.next();
            if (!this.isExpire(online,interim_times))
                continue;//直到没有超时的用户

            it.remove();

            //超时下线消息跟踪
            logger.info("BRAS[nasip="+online.getNasAddr()+",nasid="+online.getNasId()+"]:用户[user="+online.getUsername()+",session="+online.getAcctSessionId()+"]上线时间["+online.getAcctStartTime()+"]超时未收到更新消息，被自动清理。");
        }
    }

    public void updateOnline(AccountingRequest request) {
        Online online = cacheData.get(request.getAcctSessionId());
        if(online!=null){
            online.setUsername(request.getUserName());
            online.setAcctSessionId(request.getAcctSessionId());
            online.setAcctSessionTime(request.getAcctSessionTime());
            online.setAcctInputTotal(request.getAcctInputTotal());
            online.setAcctOutputTotal(request.getAcctOutputTotal());
            online.setAcctInputPackets(request.getAcctInputPackets());
            online.setAcctOutputPackets(request.getAcctOutputPackets());
        }
    }

    private boolean filterOnline(Online online, Integer groupId, String nasAddr, String nasId, String beginTime, String endTime,  String keyword) {
        if(groupId!=null&&groupId!=online.getGroupId()) {
            return false;
        }
        if(ValidateUtil.isNotEmpty(nasAddr)&&(!nasAddr.equalsIgnoreCase(online.getNasAddr())&&!nasAddr.equals(online.getNasPaddr()))) {
            return  false;
        }
        if(ValidateUtil.isNotEmpty(nasId)&&!nasId.equalsIgnoreCase(online.getNasId())) {
            return  false;
        }

        if (ValidateUtil.isNotEmpty(beginTime)) {
            if(beginTime.length() == 16){
                beginTime = beginTime+ ":00";
            }
            if( DateTimeUtil.compareSecond(online.getAcctStartTime(), beginTime)<0){
                return  false;
            }
        }

        if (ValidateUtil.isNotEmpty(endTime)) {
            if(endTime.length()==16){
                endTime = endTime + ":59";
            }
            if( DateTimeUtil.compareSecond(online.getAcctStartTime(), endTime)>0){
                return  false;
            }
        }

        if(ValidateUtil.isNotEmpty(keyword)){
            if( (ValidateUtil.isNotEmpty(online.getUsername()) && online.getUsername().toLowerCase().contains(keyword.toLowerCase())) ||
                (ValidateUtil.isNotEmpty(online.getNasAddr()) && online.getNasAddr().contains(keyword)) ||
                (ValidateUtil.isNotEmpty(online.getNasPaddr()) && online.getNasPaddr().contains(keyword)) ||
                (ValidateUtil.isNotEmpty(online.getNasId()) && online.getNasId().contains(keyword)) ||
                (ValidateUtil.isNotEmpty(online.getNasPaddr()) && online.getNasPaddr().contains(keyword)) ||
                (ValidateUtil.isNotEmpty(online.getNasPortId()) && online.getNasPortId().contains(keyword)) ||
                (ValidateUtil.isNotEmpty(online.getFramedIpaddr())&& online.getFramedIpaddr().contains(keyword))||
                (ValidateUtil.isNotEmpty(online.getMacAddr())&& online.getMacAddr().contains(keyword)) ){
                return  true;
            }else{
                return false;
            }
        }
        return true;
    }

    public PageResult<Online> queryOnlinePage(int pos, int count, Integer groupId, String nasAddr, String nasId, String beginTime, String endTime, String keyword){
        int total = 0;
        int start = pos+1;
        int end = pos +  count ;

        List<Online> copyList = new ArrayList<Online>(cacheData.values());
        List<Online> onlineList = new ArrayList<Online>();
        Comparator<Online> comp = (Online a, Online b) -> {
            return (int)DateTimeUtil.compareSecond(b.getAcctStartTime(),a.getAcctStartTime());
        };
        copyList.sort(comp);
        for (Online online : copyList) {
            if (!this.filterOnline(online, groupId, nasAddr, nasId, beginTime, endTime, keyword)) {
                continue;
            }
            else{
                total++;
                if (total >= start && total <= end) {
                    try {
                        onlineList.add(online.clone());
                    } catch (CloneNotSupportedException e) {
                        e.printStackTrace();
                    }
                }
            }

        }
        return new PageResult<Online>(pos, total, onlineList);
    }


    public  List<Online> queryOnlineByIds(String ids){
        String[] idarray = ids.split(",");
        List<Online> onlineList = new ArrayList<Online>();
        for (String sid : idarray) {
            Online online = cacheData.get(sid);
            if(online!=null){
                onlineList.add(online);
            }
        }
        return onlineList;
    }

    public int clearOnlineByFilter(Integer groupId, String nasAddr, String nasId, String beginTime, String endTime,  String keyword){
        int total = 0;
        List<Online> onlineList = new ArrayList<Online>();
        for (Iterator<Online> it = cacheData.values().iterator(); it.hasNext();)
        {
            Online online = it.next();
            if(this.filterOnline(online, groupId,   nasAddr, nasId, beginTime, endTime,  keyword)) {
                total++;
                it.remove();
            }
        }
        return total;
    }

    public int clearOnlineByFilter(String nasAddr, String nasId){
        int total = 0;
        for (Iterator<Online> it = cacheData.values().iterator(); it.hasNext();)
        {
            Online online = it.next();
            if(this.filterOnline(online, null,  nasAddr, nasId, null, null,  null)) {
                total++;
                it.remove();
            }
        }
        return total;
    }






}
