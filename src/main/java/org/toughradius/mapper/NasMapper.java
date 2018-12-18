package org.toughradius.mapper;

import org.apache.ibatis.annotations.Mapper;
import org.springframework.stereotype.Repository;
import org.toughradius.entity.Nas;

@Repository
@Mapper
public interface NasMapper {

    Nas selectByIPAddr(String ipaddr);

    Nas selectByIdentifier(String identifier);
}
