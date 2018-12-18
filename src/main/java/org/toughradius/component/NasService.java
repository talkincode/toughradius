package org.toughradius.component;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.toughradius.common.ValidateUtil;
import org.toughradius.entity.Nas;
import org.toughradius.mapper.NasMapper;

@Service
public class NasService {

	@Autowired
	private NasMapper nasMapper;

	private final static Log logger = LogFactory.getLog(NasService.class);

	public Nas findNas(String ipaddr, String identifier) throws ServiceException{
		Nas nas = null;
		if(ValidateUtil.isNotEmpty(ipaddr)&&!"0.0.0.0".equals(ipaddr)){
			nas = nasMapper.selectByIPAddr(ipaddr);
		}

		if (nas == null) {
			nas = nasMapper.selectByIdentifier(identifier);
		}

		if (nas == null) {
			throw new ServiceException(String.format("Bras设备 id=%s, ip=%s 不存在", identifier, ipaddr));
		}

		if (nas.getStatus() != null && "disabled".equals(nas.getStatus())) {
			throw new ServiceException(String.format("Bras设备 id=%s, ip=%s 已停用", identifier, ipaddr));
		}

		return nas;
	}
}
