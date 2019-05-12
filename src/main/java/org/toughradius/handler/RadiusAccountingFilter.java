package org.toughradius.handler;

import org.toughradius.common.CoderUtil;
import org.toughradius.common.ValidateCache;
import org.toughradius.component.OnlineCache;
import org.toughradius.component.SubscribeCache;
import org.toughradius.component.TicketCache;
import org.toughradius.common.DateTimeUtil;
import org.toughradius.config.RadiusConfig;
import org.toughradius.entity.*;
import org.tinyradius.packet.AccountingRequest;
import org.tinyradius.util.RadiusException;
import org.toughradius.component.RadiusStat;
import org.toughradius.component.SubscribeService;
import org.toughradius.component.Memarylogger;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.scheduling.concurrent.ThreadPoolTaskExecutor;
import org.springframework.stereotype.Component;

import java.math.BigInteger;
import java.util.Date;

@Component
public class RadiusAccountingFilter {

    @Autowired
    protected RadiusStat radiusStat;
    @Autowired
    public Memarylogger logger;
    @Autowired
    private OnlineCache onlineCache;
    @Autowired
    private TicketCache ticketCache;
    @Autowired
    private RadiusConfig radiusConfig;
    @Autowired
    private SubscribeCache subscribeCache;
    @Autowired
    private SubscribeService subscribeService;
    @Autowired
    private ThreadPoolTaskExecutor systaskExecutor;


    public final static int ACCOUNTING_STATUS_START = 1;
    public final static int ACCOUNTING_STATUS_STOP = 2;
    public final static int ACCOUNTING_STATUS_UPDATE = 3;
    public final static int ACCOUNTING_STATUS_ON = 7;
    public final static int ACCOUNTING_STATUS_OFF = 8;

    /**
     * 重复下先报文检测,每个会话每秒只能有一个
     */
    private ValidateCache onlineStopValid= new ValidateCache(3000,1);

    @Scheduled(fixedRate = 60 * 1000)
    public void  checkValidateCacheExpire(){
        onlineStopValid.clearExpire();
    }


    /**
     * RADIUS 记账处理
     * @param request
     * @param nas
     * @param user
     * @throws RadiusException
     */
    public void doFilter(AccountingRequest request, Bras nas, Subscribe user) throws RadiusException {
        int type = request.getAcctStatusType();
        if (type == ACCOUNTING_STATUS_START) {
            radiusStat.incrAcctStart();
            doStart(request,nas,user);
        }else if (type == ACCOUNTING_STATUS_UPDATE) {
            radiusStat.incrAcctUpdate();
            doUpdate(request,nas,user);
        }else if (type == ACCOUNTING_STATUS_STOP) {
            radiusStat.incrAcctStop();
            doStop(request,nas,user);
        }else if (type == ACCOUNTING_STATUS_ON) {
            radiusStat.incrAcctOn();
            doNasOn(request,nas,user);
        }else if (type == ACCOUNTING_STATUS_OFF) {
            radiusStat.incrAcctOff();
            doNasOff(request,nas,user);
        }else{
            logger.error(request.getUsername(),String.format("无效的记账请求类型 %s", type), Memarylogger.RADIUSD);
        }
    }


    /**
     * 新增在线用户
     * @param request
     * @param nas
     * @param user
     * @throws RadiusException
     */
    private void addOnline(AccountingRequest request, Bras nas, Subscribe user) throws RadiusException {
        if(user==null){
            String errMessage = String.format("用户不存在或状态未启用（记账请求） %s", request.getAcctStatusType());
            logger.error(request.getUsername(), errMessage, Memarylogger.RADIUSD);
            return;
        }
        if(onlineCache.isExist(request.getAcctSessionId())){
            logger.error(request.getUsername(),"记账报文重复", Memarylogger.RADIUSD);
            return;
        }

        int onlineCount = onlineCache.getUserOnlineNum(request.getUserName());
        if (onlineCount >= user.getActiveNum()) {
            String errMessage = String.format("用户<%s>在线超过限制(MAX=%s)(记账请求) %s",user.getSubscriber(), user.getActiveNum(),request.getAcctStatusType());
            logger.error(request.getUsername(),errMessage, Memarylogger.RADIUSD);
            return;
        }
        RadiusOnline online = new RadiusOnline();
//        online.setId(CoderUtil.randomLongId());
        online.setNodeId(user.getNodeId());
        online.setRealname(user.getRealname());
        online.setUsername(request.getUserName());
        online.setNasId(request.getIdentifier());
        online.setNasAddr(nas.getIpaddr());
        online.setNasPaddr(request.getRemoteAddr().getHostName());
        online.setSessionTimeout(request.getSessionTimeout());
        online.setFramedIpaddr(request.getFramedIpaddr());
        online.setFramedNetmask(request.getFramedNetmask());
        online.setMacAddr(request.getMacAddr());
        online.setInVlan(request.getInVlanId());
        online.setOutVlan(request.getOutVlanId());
        online.setNasPort((long) request.getNasPort());
        online.setNasClass(request.getNasClass());
        online.setNasPortId(request.getNasPortId());
        online.setServiceType(request.getServiceType());
        online.setAcctSessionId(request.getAcctSessionId());
        online.setAcctSessionTime(request.getAcctSessionTime());
        online.setAcctInputTotal(request.getAcctInputTotal());
        online.setAcctOutputTotal(request.getAcctOutputTotal());
        online.setAcctInputPackets(request.getAcctInputPackets());
        online.setAcctOutputPackets(request.getAcctOutputPackets());
        online.setAcctStartTime(request.getAcctStartTime());
        onlineCache.putOnline(online);
        subscribeCache.startSubscribeOnline(request.getUsername());
        if(radiusConfig.getTrace() ==1){
            logger.info(request.getUsername(),String.format(":: 新增用户在线信息: sessionId=%s", request.getAcctSessionId()), Memarylogger.RADIUSD);
        }

    }


    public void doStart(AccountingRequest request, Bras nas, Subscribe user) throws RadiusException {
        addOnline(request,nas,user);
    }

    public void doUpdate(AccountingRequest request, Bras nas, Subscribe user) throws RadiusException {
        systaskExecutor.execute(() -> {
            try {
                if(!onlineCache.isExist(request.getAcctSessionId())){
                    logger.print(String.format(":: 更新在线用户时用户在线记录不存在，立即新增在线信息 %s ", request.getUserName()));
                    addOnline(request,nas,user);
                    return;
                }
                //2. 更新在线数据
                onlineCache.updateOnline(request);
                if (radiusConfig.getTrace() == 1) {
                    logger.print(String.format(":: %s 结束记账更新 ", request.getUsername()));
                }
            } catch (Exception e) {
                logger.error(request.getUsername(),"记账更新错误",e, Memarylogger.RADIUSD);
            }
        });
    }

    public void doStop(AccountingRequest request, Bras nas, Subscribe user) throws RadiusException {
        systaskExecutor.execute(() -> {
            try {
                onlineStopValid.incr(request.getAcctSessionId());
                if(onlineStopValid.isOver(request.getAcctSessionId())){
                    logger.info(request.getUsername(), String.format(":: 收到重复记账下线消息:%s;", request.toString()), Memarylogger.RADIUSD);
                    return;
                }

                if(radiusConfig.getTrace()==1){
                    if(!onlineCache.isExist(request.getAcctSessionId()))
                    {
                        logger.info(request.getUsername(), String.format(":: 收到记账下线消息(但在线用户不存在):%s", request.toString()), Memarylogger.RADIUSD);
                        return;
                    }
                    logger.info(request.getUsername(),":: 开始记账下线处理:" + request.toString(), Memarylogger.RADIUSD);
                }

                RadiusOnline online = onlineCache.removeOnline(request.getAcctSessionId());
                subscribeCache.stopSubscribeOnline(online.getUsername());
                //新增上网日志
                RadiusTicket radiusTicket = new RadiusTicket();
                radiusTicket.setId(CoderUtil.randomLongId());
                radiusTicket.setNodeId(user.getNodeId());
                radiusTicket.setUsername(request.getUserName());
                radiusTicket.setNasId(request.getIdentifier());
                radiusTicket.setNasAddr(nas.getIpaddr());
                radiusTicket.setNasPaddr(request.getRemoteAddr().getHostName());
                radiusTicket.setAcctSessionTime(request.getAcctSessionTime());
                radiusTicket.setSessionTimeout(request.getSessionTimeout());
                radiusTicket.setFramedIpaddr(request.getFramedIpaddr());
                radiusTicket.setFramedNetmask(request.getFramedNetmask());
                radiusTicket.setMacAddr(request.getMacAddr());
                radiusTicket.setInVlan(request.getInVlanId());
                radiusTicket.setOutVlan(request.getOutVlanId());
                radiusTicket.setNasPort((long) request.getNasPort());
                radiusTicket.setNasClass(request.getNasClass());
                radiusTicket.setNasPortId(request.getNasPortId());
                radiusTicket.setNasPortType(0);
                radiusTicket.setServiceType(request.getServiceType());
                radiusTicket.setAcctSessionId(request.getAcctSessionId());
                radiusTicket.setAcctInputTotal(request.getAcctInputTotal());
                radiusTicket.setAcctOutputTotal(request.getAcctOutputTotal());
                radiusTicket.setAcctInputPackets(request.getAcctInputPackets());
                radiusTicket.setAcctOutputPackets(request.getAcctOutputPackets());
                radiusTicket.setAcctStopTime(new Date());
                radiusTicket.setAcctStartTime(DateTimeUtil.toDate(request.getAcctStartTime()));
                ticketCache.addTicket(radiusTicket);
                //删除在线数据
                if(radiusConfig.getTrace()==1)
                    logger.print("新增上网日志");
            }catch (Exception e) {
                logger.error(request.getUsername(), "用户下线处理错误", e, Memarylogger.RADIUSD);
            }
        });
    }

    public void doNasOn(AccountingRequest request, Bras nas, Subscribe user) throws RadiusException {
        if(radiusConfig.getTrace() ==1){
            logger.info(String.format(":: NAS <%s %s> 记账启用 ", nas.getIdentifier(), nas.getIpaddr()), Memarylogger.RADIUSD);
        }
        onlineCache.clearOnlineByFilter(nas.getIpaddr(),nas.getIdentifier());
    }

    public void doNasOff(AccountingRequest request, Bras nas, Subscribe user) throws RadiusException {
        if(radiusConfig.getTrace() ==1){
            logger.info(String.format(":: NAS <%s %s> 记账关闭 ", nas.getIdentifier(), nas.getIpaddr()), Memarylogger.RADIUSD);
        }
        onlineCache.clearOnlineByFilter(nas.getIpaddr(),nas.getIdentifier());
    }


}
