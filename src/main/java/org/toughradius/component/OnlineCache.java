package org.toughradius.component;
import org.toughradius.config.RadiusConfig;
import org.toughradius.entity.Bras;
import org.tinyradius.packet.AccountingRequest;
import org.toughradius.common.DateTimeUtil;
import org.tinyradius.packet.CoaRequest;
import org.tinyradius.packet.RadiusPacket;
import org.tinyradius.util.RadiusClient;
import org.tinyradius.util.RadiusException;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.concurrent.ThreadPoolTaskExecutor;
import org.springframework.stereotype.Service;
import org.toughradius.common.PageResult;
import org.toughradius.common.ValidateUtil;
import org.toughradius.entity.RadiusOnline;
import org.toughradius.entity.Subscribe;
import java.io.*;
import java.util.*;

@Service
public class OnlineCache {

    private final static  HashMap<String,RadiusOnline> cacheData = new HashMap<String,RadiusOnline>();
    @Autowired
    private Syslogger logger;

    @Autowired
    private RadiusConfig radiusConfig;


    @Autowired
    private BrasService brasService;

    @Autowired
    private SubscribeCache subscribeCache;

    @Autowired
    private ThreadPoolTaskExecutor taskExecutor;


    public HashMap<String,RadiusOnline> getCacheData(){
        return cacheData;
    }

    private Object flowStatLock = new Object();


    public int size()
    {
        return cacheData.size();
    }

    public List<RadiusOnline> getOnlineList()
    {
        return new ArrayList<RadiusOnline>(Collections.unmodifiableCollection(cacheData.values()));
    }

    public List<RadiusOnline> list(String nodeId)
    {
        List<RadiusOnline> onlineList = new ArrayList<RadiusOnline>();
        synchronized (cacheData)
        {
            for (RadiusOnline online : cacheData.values()) {
                if (online != null && online.getNodeId().equals(nodeId))
                    onlineList.add(online);
            }
        }
        return onlineList;
    }

    public List<RadiusOnline> queryNoAmountOnline( )
    {
        List<RadiusOnline> onlineList = new ArrayList<RadiusOnline>();
        synchronized (cacheData){
            for (RadiusOnline online : cacheData.values()) {
                if (online.getUnLockFlag() == RadiusOnline.AMOUNT_NOT_ENOUGH)
                    onlineList.add(online);
            }
        }
        return onlineList;
    }

    public boolean isExist(String sessionId)
    {
        synchronized (cacheData){
            return cacheData.containsKey(sessionId);
        }
    }

    public RadiusOnline getOnline(String sessionId)
    {
        synchronized (cacheData){
            return cacheData.get(sessionId);
        }
    }

    public boolean isOnline(String userName)
    {
        boolean isTcRadiusOnline = false;
        synchronized (cacheData){
            for (RadiusOnline online : cacheData.values()) {
                if (userName.equalsIgnoreCase(online.getUsername())) {
                    isTcRadiusOnline = true;
                    break;
                }
            }
        }
        return isTcRadiusOnline;
    }

    public RadiusOnline getFirstOnline(String userName)
    {
        synchronized (cacheData){
            for (RadiusOnline online : cacheData.values()) {
                if (userName.equalsIgnoreCase(online.getUsername())) {
                    return online;
                }
            }
        }
        return null;
    }


    public RadiusOnline getLastOnline(String userName)
    {
        RadiusOnline online =null;
        synchronized (cacheData){
            for (RadiusOnline _online : cacheData.values()) {
                if (userName.equalsIgnoreCase(_online.getUsername())) {
                    online = _online;
                    break;
                }
            }
        }
        return online;
    }


    /**
     * 异步批量下线
     * @param ids
     */
    public void unlockOnlines(List<String> ids)
    {
        for(String sessionid : ids){
            asyncUnlockOnline(sessionid);
        }
    }

    /**
     * 异步批量下线
     * @param ids
     */
    public void unlockOnlines(String ids)
    {
        for(String sessionid : ids.split(",")){
            asyncUnlockOnline(sessionid);
        }
    }

    /**
     * 单个下线
     * @param sessionId
     */
    public boolean unlockOnline(String sessionId)
    {
        RadiusOnline online = getOnline(sessionId);
        try {
            if(online==null){
                logger.error("发送下线失败,无在线信息",Syslogger.RADIUSD);
                return false;
            }
            Bras bras = brasService.findBras(online.getNasAddr(),null, online.getNasId());
            RadiusClient cli = new RadiusClient(online.getNasPaddr(),bras.getSecret());
            CoaRequest dmreq = new CoaRequest(RadiusPacket.DISCONNECT_REQUEST);
            dmreq.addAttribute("User-Name",online.getUsername());
            dmreq.addAttribute("Acct-Session-Id",online.getAcctSessionId());
            dmreq.addAttribute("NAS-IP-Address",online.getNasAddr());
            logger.info(online.getUsername(), String.format("发送下线请求 %s", dmreq.toLineString()),Syslogger.RADIUSD);
            RadiusPacket dmrep = cli.communicate(dmreq,bras.getCoaPort());
            logger.info(online.getUsername(), String.format("接收到下线响应 %s", dmrep.toLineString()),Syslogger.RADIUSD);
            return dmrep.getPacketType() == RadiusPacket.DISCONNECT_ACK;
        } catch (ServiceException | IOException | RadiusException e) {
            logger.error(online.getUsername(),"发送下线失败",e,Syslogger.RADIUSD);
            removeOnline(sessionId);
            return false;
        }
    }

    public void asyncUnlockOnline(String sessionid){
        RadiusOnline online = getOnline(sessionid);
        if(online==null){
            logger.error("发送下线失败,无在线信息",Syslogger.RADIUSD);
            return;
        }
        taskExecutor.execute(()->{
            try {
                Bras bras = brasService.findBras(online.getNasAddr(),null,online.getNasId());
                if(bras==null){
                    logger.error(online.getUsername(),
                            String.format("发送下线失败,查找BRAS失败（nasid=%s,nasip=%s）",
                                    online.getNasId(), online.getNasAddr()),Syslogger.RADIUSD);
                    return;
                }
                RadiusClient cli = new RadiusClient(online.getNasPaddr(),bras.getSecret());
                CoaRequest dmreq = new CoaRequest(RadiusPacket.DISCONNECT_REQUEST);
                dmreq.addAttribute("User-Name",online.getUsername());
                dmreq.addAttribute("Acct-Session-Id",online.getAcctSessionId());
                if(ValidateUtil.isNotEmpty(online.getNasAddr())&&!online.getNasAddr().equals("0.0.0.0"))
                    dmreq.addAttribute("NAS-IP-Address",online.getNasAddr());
                logger.info(online.getUsername(), String.format("发送下线请求 %s", dmreq.toLineString()),Syslogger.RADIUSD);
                RadiusPacket dmrep = cli.communicate(dmreq,bras.getCoaPort());
                logger.info(online.getUsername(), String.format("接收到下线响应 %s", dmrep.toLineString()),Syslogger.RADIUSD);
            } catch (ServiceException | IOException | RadiusException e) {
                logger.error(online.getUsername(),"发送下线失败",e,Syslogger.RADIUSD);
                removeOnline(sessionid);
            }
        });
    }


    public List<RadiusOnline> getOnlineByUserName(String userName)
    {
        List<RadiusOnline> onlineList = new ArrayList<RadiusOnline>();
        synchronized (cacheData){
            for (RadiusOnline online : cacheData.values()) {
                if (userName.equalsIgnoreCase(online.getUsername())) {
                    onlineList.add(online);
                }
            }
        }
        return onlineList;
    }

    /** 用户上线 */
    public void putOnline(RadiusOnline online)
    {
        synchronized (cacheData)
        {
            String key = online.getAcctSessionId();
            cacheData.put(key, online);
        }
    }

    /** 一个用户下线 */
    public RadiusOnline removeOnline(String sessionId)
    {
        synchronized (cacheData){
            return (RadiusOnline) cacheData.remove(sessionId);
        }
    }

    /** 设置解锁标记 */
    public void setUnLock(String sessionId,int  flag)
    {
        synchronized (cacheData){
            RadiusOnline online = (RadiusOnline) cacheData.get(sessionId);
            if(online!=null){
                online.setUnLockFlag(flag);
            }
        }
    }

    /** BAS所有在线用户下线 */
    public List<RadiusOnline> removeAllOnline(String nasAddr, String nasId)
    {
        List<RadiusOnline> onlineList = new ArrayList<RadiusOnline>();
        synchronized (cacheData)
        {
            for (Iterator<RadiusOnline> it = cacheData.values().iterator(); it.hasNext();)
            {
                RadiusOnline online = it.next();
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
        synchronized (cacheData){
            for (RadiusOnline online : cacheData.values()) {
                if (userName.equalsIgnoreCase(online.getUsername()))
                    onlineNum++;
            }
        }
        return onlineNum;
    }

    /** 查询上网帐号并发数,要求MAC地址不相等 */
    public int getUserOnlineNum(String userName, String macAddr)
    {
        int onlineNum = 0;
        synchronized (cacheData){
            for (RadiusOnline online : cacheData.values()) {
                if (userName.equalsIgnoreCase(online.getUsername()) && !macAddr.equalsIgnoreCase(online.getMacAddr()))
                    onlineNum++;
            }
        }
        return onlineNum;
    }

    /**
     * 间隔 interim_times 不更新的为过期
     * @param online
     * @param interim_times
     * @return
     */
    private  boolean isExpire(RadiusOnline online, int interim_times)
    {
        String curTime = DateTimeUtil.getDateTimeString();
        String acctStart =  online.getAcctStartTime();
        int second = DateTimeUtil.compareSecond(curTime,acctStart);
        if (second > (online.getAcctSessionTime()+interim_times+120))
            return true;
        else
            return false;
    }


    public void clearOvertimeTcRadiusOnline( int interim_times)
    {
        synchronized (cacheData){
            for (Iterator<RadiusOnline> it = cacheData.values().iterator(); it.hasNext();)
            {
                RadiusOnline online = it.next();
                if (!this.isExpire(online,interim_times))
                    continue;//直到没有超时的用户

                it.remove();

                //超时下线消息跟踪
                logger.info(online.getUsername(),"BRAS[nasip="+online.getNasAddr()+",nasid="+online.getNasId()+"]:用户[user="+online.getUsername()+",session="+online.getAcctSessionId()+"]上线时间["+online.getAcctStartTime()+"]超时未收到更新消息，被自动清理。");
            }
        }
    }

    public void unlockExpireTcRadiusOnline()
    {

        List ids = new ArrayList();
        synchronized (cacheData){
            for (Iterator<RadiusOnline> it = cacheData.values().iterator(); it.hasNext();)
            {
                RadiusOnline online = it.next();
                Subscribe user = subscribeCache.findSubscribe(online.getUsername());
                if(user == null)
                    continue;

                if(DateTimeUtil.compareSecond(new Date(),user.getExpireTime())>0){
                    ids.add(online.getAcctSessionId());
                }else if("flow".equals(user.getBillType()) && user.getFlowAmount().intValue() <= 0){
                    ids.add(online.getAcctSessionId());
                }

                if(ids.size()>=32){
                    List idscopy = new ArrayList(ids);
                    this.unlockOnlines(idscopy);
                    ids.clear();
                }

                //超时下线消息跟踪
                logger.info(online.getUsername(),"BRAS[nasip="+online.getNasAddr()+",nasid="+online.getNasId()+"]:用户[user="+online.getUsername()+",session="+online.getAcctSessionId()+"]已经过期/或流量不足，即将发送下线指令。");
            }
        }

        this.unlockOnlines(ids);
    }

    public void updateOnline(AccountingRequest request) {
        synchronized (cacheData){
            RadiusOnline online = cacheData.get(request.getAcctSessionId());
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
    }


    private boolean filterOnline(RadiusOnline online, String nodeId, String areaId, Integer invlan, Integer outVlan, String nasAddr, String nasId, String beginTime, String endTime, String keyword) {
        if(ValidateUtil.isNotEmpty(nodeId)&&!nodeId.equalsIgnoreCase(online.getNodeId().toString())) {
            return false;
        }
        if(ValidateUtil.isNotEmpty(areaId)&&!areaId.equalsIgnoreCase(online.getAreaId().toString())) {
            return false;
        }
        if(ValidateUtil.isNotEmpty(nasAddr)&&(!nodeId.equalsIgnoreCase(online.getNasAddr())&&!nodeId.equals(online.getNasPaddr()))) {
            return  false;
        }
        if(ValidateUtil.isNotEmpty(nasId)&&!nasId.equalsIgnoreCase(online.getNasId())) {
            return  false;
        }

        if(invlan!=null&&!invlan.equals(online.getInVlan())) {
            return  false;
        }
        if(outVlan!=null&&!outVlan.equals(online.getOutVlan())) {
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
                (ValidateUtil.isNotEmpty(online.getRealname())&& online.getRealname().contains(keyword))||
                (ValidateUtil.isNotEmpty(online.getMacAddr())&& online.getMacAddr().contains(keyword)) ){
                return  true;
            }else{
                return false;
            }
        }
        return true;
    }

    public PageResult<RadiusOnline> queryOnlinePage(int pos, int count, String nodeId,
                                                    String areaId, Integer invlan, Integer outVlan, String nasAddr, String nasId,
                                                    String beginTime, String endTime, String keyword, String sort){
        int total = 0;
        int start = pos+1;
        int end = pos +  count ;

        synchronized (cacheData){
            List<RadiusOnline> copyList = new ArrayList<RadiusOnline>(cacheData.values());
            List<RadiusOnline> onlineList = new ArrayList<RadiusOnline>();
            Comparator<RadiusOnline> comp = (RadiusOnline a, RadiusOnline b) -> {
                if("acctInputTotal_asc".equals(sort)){
                    return a.getAcctInputTotal().compareTo(b.getAcctInputTotal());
                }else if("acctInputTotal_desc".equals(sort)){
                    return b.getAcctInputTotal().compareTo(a.getAcctInputTotal());
                }else if("acctOutputTotal_asc".equals(sort)){
                    return a.getAcctOutputTotal().compareTo(b.getAcctOutputTotal());
                }else if("acctOutputTotal_desc".equals(sort)){
                    return b.getAcctOutputTotal().compareTo(a.getAcctOutputTotal());
                }else if("acctStartTime_asc".equals(sort)){
                    return (int)DateTimeUtil.compareSecond(a.getAcctStartTime(),b.getAcctStartTime());
                }else if("acctStartTime_desc".equals(sort)){
                    return (int)DateTimeUtil.compareSecond(b.getAcctStartTime(),a.getAcctStartTime());
                }else{
                    return (int)DateTimeUtil.compareSecond(b.getAcctStartTime(),a.getAcctStartTime());
                }
            };
            copyList.sort(comp);
            for (RadiusOnline online : copyList) {
                if (!this.filterOnline(online, nodeId, areaId, invlan,outVlan, nasAddr, nasId, beginTime, endTime, keyword)) {
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
            return new PageResult<RadiusOnline>(pos, total, onlineList);
        }
    }


    public  List<RadiusOnline> queryOnlineByIds(String ids){
        String[] idarray = ids.split(",");
        List<RadiusOnline> onlineList = new ArrayList<RadiusOnline>();
        synchronized (cacheData){
            for (String sid : idarray) {
                RadiusOnline online = cacheData.get(sid);
                if(online!=null){
                    onlineList.add(online);
                }
            }
        }
        return onlineList;
    }

    public int clearOnlineByFilter(String nodeId, String areaId, Integer invlan,Integer outVlan,String nasAddr, String nasId, String beginTime, String endTime,  String keyword){
        int total = 0;
        List<RadiusOnline> onlineList = new ArrayList<RadiusOnline>();
        synchronized (cacheData){
            for (Iterator<RadiusOnline> it = cacheData.values().iterator(); it.hasNext();)
            {
                RadiusOnline online = it.next();
                if(this.filterOnline(online, nodeId, areaId,  invlan, outVlan, nasAddr, nasId, beginTime, endTime,  keyword)) {
                    total++;
                    it.remove();
                }
            }
        }
        return total;
    }

    public int clearOnlineByFilter(String nasAddr, String nasId){
        int total = 0;
        synchronized (cacheData){
            for (Iterator<RadiusOnline> it = cacheData.values().iterator(); it.hasNext();)
            {
                RadiusOnline online = it.next();
                if(this.filterOnline(online, null, null, null,null, nasAddr, nasId, null, null,  null)) {
                    total++;
                    it.remove();
                }
            }
        }
        return total;
    }






}
