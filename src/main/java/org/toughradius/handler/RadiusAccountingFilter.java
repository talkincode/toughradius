package org.toughradius.handler;

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
import org.toughradius.component.Syslogger;
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
    public Syslogger logger;
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
    private ThreadPoolTaskExecutor taskExecutor;


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
            logger.error(request.getUsername(),String.format("无效的记账请求类型 %s", type),Syslogger.RADIUSD);
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
            logger.error(request.getUsername(), errMessage,Syslogger.RADIUSD);
            return;
        }
        if(onlineCache.isExist(request.getAcctSessionId())){
            logger.error(request.getUsername(),"记账报文重复",Syslogger.RADIUSD);
        }

        int onlineCount = onlineCache.getUserOnlineNum(request.getUserName());
        if (onlineCount >= user.getActiveNum()) {
            String errMessage = String.format("用户<%s>在线超过限制(MAX=%s)(记账请求) %s",user.getSubscriber(), user.getActiveNum(),request.getAcctStatusType());
            logger.error(request.getUsername(),errMessage,Syslogger.RADIUSD);
            return;
        }
        SubscribeBill billData = subscribeCache.getBillData(request.getUserName());
        SubscribeBill billData2 = new SubscribeBill();
        Subscribe subscribe = subscribeCache.findSubscribe(request.getUsername());
        billData2.setBillType(subscribe.getBillType());
        RadiusOnline online = new RadiusOnline();
        online.setNodeId(user.getNodeId());
        online.setAreaId(user.getAreaId());
        online.setRealname(user.getRealname());
        online.setUsername(request.getUserName());
        if (billData==null){
            online.setBillType(billData2.getBillType());
        }else {
            online.setBillType(billData.getBillType());
        }
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
            logger.info(request.getUsername(),String.format(":: 新增用户在线信息: sessionId=%s", request.getAcctSessionId()),Syslogger.RADIUSD);
        }

    }

    /**
     * 记账扣除流量, 必须再在线数据更新或删除前完成
     * @param request
     */
    private void radiusBilling(AccountingRequest request) {
        try{
            RadiusOnline online = onlineCache.getOnline(request.getAcctSessionId());
            if (online!=null &&"flow".equals(online.getBillType())) {
                if (radiusConfig.getTrace() == 1) {
                    logger.info(request.getUsername(), "开始用户记账扣费处理",Syslogger.RADIUSD);
                }

                //获取计费类型和用户剩余流量
                SubscribeBill billData = subscribeCache.getBillData(request.getUserName());
                if (billData == null) {
                    onlineCache.setUnLock(request.getAcctSessionId(), RadiusOnline.USER_NOT_EXIST);
                    logger.error(request.getUsername(), "处理记账扣费时用户不存在",Syslogger.RADIUSD);
                    return;
                }
                Long outputTotal = online.getAcctOutputTotal();
                Long inputTotal = online.getAcctInputTotal();
                long useFlows = 0;
                if (outputTotal != null) {
                    long curr_output_total = request.getAcctOutputTotal();
                    useFlows = curr_output_total - outputTotal;

                    if (radiusConfig.isBillInput() && inputTotal != null && inputTotal > 0) {
                        long curr_input_total = request.getAcctInputTotal();
                        useFlows += curr_input_total - inputTotal;
                    }

                    if (radiusConfig.isBillBackFlow()) {
                        long inpkts = request.getAcctInputPackets() - online.getAcctInputPackets();
                        long outpkts = request.getAcctOutputPackets() - online.getAcctOutputPackets();
                        long flows = (inpkts + outpkts) * 80;
                        useFlows += flows;
                    }

                    if (useFlows <= 0) {
                        useFlows = 0;
                    }
                    if (useFlows > 0) {
                        if (billData.getFlowAmount().longValue() > useFlows) {
                            subscribeService.updateFlowAmountByUsername(request.getUserName(),  BigInteger.valueOf(billData.getFlowAmount().longValue() - useFlows));
                        } else {
                            subscribeService.updateFlowAmountByUsername(request.getUserName(), BigInteger.ZERO);
                            onlineCache.setUnLock(online.getAcctSessionId(), RadiusOnline.AMOUNT_NOT_ENOUGH);
                        }
                    }
                }
                if (radiusConfig.getTrace() == 1) {
                    logger.info(request.getUsername(), ":: 完成用户 %s 记账扣费处理",Syslogger.RADIUSD);
                }
            }
        }catch (Exception e){
            logger.error(request.getUsername(),  "用户记账扣费处理失败", e,Syslogger.RADIUSD);
        }
    }

    public void doStart(AccountingRequest request, Bras nas, Subscribe user) throws RadiusException {
        addOnline(request,nas,user);
    }

    public void doUpdate(AccountingRequest request, Bras nas, Subscribe user) throws RadiusException {
        taskExecutor.execute(() -> {
            try {
                if(!onlineCache.isExist(request.getAcctSessionId())){
                    logger.print(String.format(":: 更新在线用户时用户在线记录不存在，立即新增在线信息 %s ", request.getUserName()));
                    addOnline(request,nas,user);
                    return;
                }
                //1. 进行流量扣费操作
                radiusBilling(request);
                //2. 更新在线数据
                onlineCache.updateOnline(request);
                if (radiusConfig.getTrace() == 1) {
                    logger.print(String.format(":: %s 结束记账更新 ", request.getUsername()));
                }
            } catch (Exception e) {
                logger.error(request.getUsername(),"记账更新错误",e,Syslogger.RADIUSD);
            }
        });
    }

    public void doStop(AccountingRequest request, Bras nas, Subscribe user) throws RadiusException {
        taskExecutor.execute(() -> {
            try {
                onlineStopValid.incr(request.getAcctSessionId());
                if(onlineStopValid.isOver(request.getAcctSessionId())){
                    logger.info(request.getUsername(), String.format(":: 收到重复记账下线消息:%s;", request.toString()),Syslogger.RADIUSD);
                    return;
                }

                if(radiusConfig.getTrace()==1){
                    if(!onlineCache.isExist(request.getAcctSessionId()))
                    {
                        logger.info(request.getUsername(), String.format(":: 收到记账下线消息(但在线用户不存在):%s", request.toString()),Syslogger.RADIUSD);
                        return;
                    }
                    logger.info(request.getUsername(),":: 开始记账下线处理:" + request.toString(),Syslogger.RADIUSD);
                }

                RadiusOnline online = onlineCache.removeOnline(request.getAcctSessionId());
                subscribeCache.stopSubscribeOnline(online.getUsername());
                //删除在线数据
                if(radiusConfig.getTrace()==1)
                    logger.print("删除在线用户缓存");

                //执行流量扣费
                radiusBilling(request);
                //新增上网日志
                int nodeId = 0;
                int areaId = 0;
                if(user!=null&&!user.getSubscriber().substring(0,2).equals("ls")){
                    nodeId = user.getNodeId();
                    areaId = user.getAreaId();
                }
                if(online!=null&&!user.getSubscriber().substring(0,2).equals("ls")){
                    nodeId = online.getNodeId();
                    areaId = online.getAreaId();
                }
                RadiusTicket radiusTicket = new RadiusTicket();
                radiusTicket.setNodeId(nodeId);
                radiusTicket.setAreaId(areaId);
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
                logger.error(request.getUsername(), "用户下线处理错误", e,Syslogger.RADIUSD);
            }
        });
    }

    public void doNasOn(AccountingRequest request, Bras nas, Subscribe user) throws RadiusException {
        if(radiusConfig.getTrace() ==1){
            logger.info(String.format(":: NAS <%s %s> 记账启用 ", nas.getIdentifier(), nas.getIpaddr()),Syslogger.RADIUSD);
        }
        onlineCache.clearOnlineByFilter(nas.getIpaddr(),nas.getIdentifier());
    }

    public void doNasOff(AccountingRequest request, Bras nas, Subscribe user) throws RadiusException {
        if(radiusConfig.getTrace() ==1){
            logger.info(String.format(":: NAS <%s %s> 记账关闭 ", nas.getIdentifier(), nas.getIpaddr()),Syslogger.RADIUSD);
        }
        onlineCache.clearOnlineByFilter(nas.getIpaddr(),nas.getIdentifier());
    }


}
