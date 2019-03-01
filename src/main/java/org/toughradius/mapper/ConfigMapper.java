package org.toughradius.mapper;

import org.toughradius.entity.Config;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.springframework.stereotype.Repository;


@Repository
@Mapper
public interface ConfigMapper {

    Integer getInterimTimes();

    Integer getIsCheckPwd();

    Config findConfig(@Param(value = "type") String type, @Param(value = "name") String name);

}