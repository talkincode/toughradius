package org.toughradius.handler;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.concurrent.ThreadPoolTaskExecutor;
import org.springframework.stereotype.Component;
import org.tinyradius.packet.AccountingRequest;
import org.tinyradius.util.RadiusException;
import org.toughradius.common.DateTimeUtil;
import org.toughradius.component.*;
import org.toughradius.config.RadiusConfig;
import org.toughradius.entity.Nas;
import org.toughradius.entity.Online;
import org.toughradius.entity.Ticket;
import org.toughradius.entity.User;

import java.math.BigInteger;
import java.util.Date;

@Component
public class RadiusAccountingFilter {

    private final static Log logger = LogFactory.getLog(RadiusAuthHandler.class);

    @Autowired
    protected RadiusStat radiusStat;

    @Autowired
    private OnlineCache onlineCache;
    @Autowired
    private TicketCache ticketCache;
    @Autowired
    private RadiusConfig radiusConfig;
    @Autowired
    private UserCache subscribeCache;
    @Autowired
    private UserService subscribeService;
    @Autowired
    private ThreadPoolTaskExecutor taskExecutor;


    public final static int ACCOUNTING_STATUS_START = 1;
    public final static int ACCOUNTING_STATUS_STOP = 2;
    public final static int ACCOUNTING_STATUS_UPDATE = 3;
    public final static int ACCOUNTING_STATUS_ON = 7;
    public final static int ACCOUNTING_STATUS_OFF = 8;

    /**
     * RADIUS 记账处理
     * @param request
     * @param nas
     * @param user
     * @throws RadiusException
     */
    public void doFilter(AccountingRequest request, Nas nas, User user) throws RadiusException {
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
            logger.error(String.format("无效的记账请求类型 %s", type));
        }
    }


    /**
     * 新增在线用户
     * @param request
     * @param nas
     * @param user
     * @throws RadiusException
     */
    private void addOnline(AccountingRequest request,Nas nas, User user) throws RadiusException {
        if(request.getAcctStatusType() == AccountingRequest.ACCT_STATUS_TYPE_START ){
            if(user==null){
                String errMessage = String.format("用户不存在或状态未启用（记账请求） %s", request.getAcctStatusType());
                logger.error(errMessage);
                return;
            }
            if(onlineCache.isExist(request.getAcctSessionId())){
                logger.error("记账报文重复");
            }

            int onlineCount = onlineCache.getUserOnlineNum(request.getUserName());
            if (onlineCount >= user.getOnlineNum()) {
                String errMessage = String.format("用户<%s>在线超过限制(MAX=%s)(记账请求) %s",user.getUsername(), user.getOnlineNum());
                logger.error(errMessage);
                return;
            }
            Online online = new Online();
            online.setGroupId(user.getGroupId());
            online.setUsername(request.getUserName());
            online.setBillType(user.getBillType());
            online.setNasId(request.getIdentifier());
            online.setNasAddr(nas.getIpaddr());
            online.setNasPaddr(request.getRemoteAddr().getHostName());
            online.setSessionTimeout(request.getSessionTimeout());
            online.setFramedIpaddr(request.getFramedIpaddr());
            online.setFramedNetmask(request.getFramedNetmask());
            online.setMacAddr(request.getMacAddr());
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
            if(radiusConfig.getTrace() ==1){
                logger.info(String.format(":: 新增用户在线信息: sessionId=%s", request.getAcctSessionId()));
            }
        }
    }

    /**
     * 记账扣除流量, 必须再在线数据更新或删除前完成
     * @param request
     */
    private void radiusBilling(AccountingRequest request) {
        try{
            Online online = onlineCache.getOnline(request.getAcctSessionId());
            if (online!=null &&"flow".equals(online.getBillType())) {
                if (radiusConfig.getTrace() == 1) {
                    logger.info("开始用户 %s 记账扣费处理");
                }

                //获取计费类型和用户剩余流量
                BigInteger flowAmount = subscribeCache.getFlowAmount(request.getUserName());
                if (flowAmount == null ) {
                    onlineCache.setUnLock(request.getAcctSessionId(), Online.USER_NOT_EXIST);
                    logger.error("处理记账扣费时用户不存在");
                    return;
                }
                Long outputTotal = online.getAcctOutputTotal();
                Long inputTotal = online.getAcctInputTotal();
                long useFlows = 0;
                if (outputTotal != null && outputTotal != 0) {
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
                        if (flowAmount.longValue() > useFlows) {
                            subscribeService.updateFlowAmountByUsername(request.getUserName(), (flowAmount.longValue() - (int) useFlows));
                        } else {
                            subscribeService.updateFlowAmountByUsername(request.getUserName(), 0L);
                            onlineCache.setUnLock(request.getAcctSessionId(), Online.AMOUNT_NOT_ENOUGH);
                        }
                    }
                }
                if (radiusConfig.getTrace() == 1) {
                    logger.info(":: 结束用户 %s 记账扣费处理");
                }
            }
        }catch (Exception e){
            logger.error( "用户记账扣费处理失败", e);
        }
    }

    public void doStart(AccountingRequest request, Nas nas, User user) throws RadiusException {
        addOnline(request,nas,user);
    }

    public void doUpdate(AccountingRequest request, Nas nas, User user) throws RadiusException {
        if(!onlineCache.isExist(request.getAcctSessionId())){
            addOnline(request,nas,user);
        }

        taskExecutor.execute(() -> {
            try {
                //1. 进行流量扣费操作
                radiusBilling(request);
                //2. 更新在线数据
                onlineCache.updateOnline(request);
                if (radiusConfig.getTrace() == 1) {
                    logger.info(":: 结束记账更新 ");
                }
            } catch (Exception e) {
                logger.error("记账更新错误",e);
            }
        });

        if(radiusConfig.getTrace()==1)
            logger.info(":: 更新在线用户 %s ");
    }

    public void doStop(AccountingRequest request, Nas nas, User user) throws RadiusException {
        if(radiusConfig.getTrace()==1){
            logger.info(":: 开始记账下线处理:" + request.toString());
        }
        taskExecutor.execute(() -> {
            try {

                //执行流量扣费
                radiusBilling(request);
                //新增上网日志
                int groupId = 0;
                if(user!=null){
                    groupId = user.getGroupId();
                }
                Online online = onlineCache.getOnline(request.getAcctSessionId());
                if(online!=null){
                    groupId = online.getGroupId();
                }
                Ticket radiusTicket = new Ticket();
                radiusTicket.setGroupId(groupId);
                radiusTicket.setUsername(request.getUserName());
                radiusTicket.setNasId(request.getIdentifier());
                radiusTicket.setNasAddr(nas.getIpaddr());
                radiusTicket.setNasPaddr(request.getRemoteAddr().getHostName());
                radiusTicket.setAcctSessionTime(request.getAcctSessionTime());
                radiusTicket.setSessionTimeout(request.getSessionTimeout());
                radiusTicket.setFramedIpaddr(request.getFramedIpaddr());
                radiusTicket.setFramedNetmask(request.getFramedNetmask());
                radiusTicket.setMacAddr(request.getMacAddr());
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
                onlineCache.removeOnline(request.getAcctSessionId());
                if(radiusConfig.getTrace()==1)
                    logger.info("删除在线用户缓存");
            }catch (Exception e) {
                logger.error("用户下线处理错误", e);
            }
        });
        if(radiusConfig.getTrace()==1)
            logger.info(":: 结束用户下线处理");
    }

    public void doNasOn(AccountingRequest request, Nas nas, User user) throws RadiusException {
        if(radiusConfig.getTrace() ==1){
            logger.info(String.format(":: NAS <%s %s> 记账启用 ", nas.getIdentifier(), nas.getIpaddr()));
        }
        onlineCache.clearOnlineByFilter(nas.getIpaddr(),nas.getIdentifier());
    }

    public void doNasOff(AccountingRequest request, Nas nas, User user) throws RadiusException {
        if(radiusConfig.getTrace() ==1){
            logger.info(String.format(":: NAS <%s %s> 记账关闭 ", nas.getIdentifier(), nas.getIpaddr()));
        }
        onlineCache.clearOnlineByFilter(nas.getIpaddr(),nas.getIdentifier());
    }


}
