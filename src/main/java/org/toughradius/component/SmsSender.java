package org.toughradius.component;

import com.github.qcloudsms.SmsMultiSender;
import com.github.qcloudsms.SmsMultiSenderResult;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.Date;


@Component
public class SmsSender {

    @Autowired
    private ConfigService configService;

    @Autowired
    private Memarylogger logger;

    /**
     * 发送腾讯云短信
     * @param phone
     * @param content
     * @return
     */
    public  int sendQcloudSms(String phone,String content)  {

        int appid = Integer.valueOf(configService.getStringValue(ConfigService.SMS_MODULE,ConfigService.SMS_APPID));
        String appkey = configService.getStringValue(ConfigService.SMS_MODULE,ConfigService.SMS_APPKEY);
        SmsMultiSender msender = new SmsMultiSender(appid, appkey);
        try {
            SmsMultiSenderResult result=msender.send(0,"86", new String[]{phone}, content, "","");
            return result.result;
        } catch (Exception e) {
            logger.error("腾讯云短信发送失败",e,Memarylogger.SYSTEM);
        }
        return  999;
    }

    public static class SmscodeCounter{
        private String phone;
        private String smscode;
        private long sendtime;


        public SmscodeCounter(String phone, String smscode) {
            this.phone = phone;
            this.smscode = smscode;
            this.sendtime = new Date().getTime();
        }

        public String getPhone() {
            return phone;
        }

        public void setPhone(String phone) {
            this.phone = phone;
        }

        public String getSmscode() {
            return smscode;
        }

        public void setSmscode(String smscode) {
            this.smscode = smscode;
        }

        public long getSendtime() {
            return sendtime;
        }

        public void setSendtime(long sendtime) {
            this.sendtime = sendtime;
        }
    }
}
