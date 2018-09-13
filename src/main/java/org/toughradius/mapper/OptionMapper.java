package org.toughradius.mapper;


import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.springframework.stereotype.Repository;
import org.toughradius.entity.Option;

@Repository
@Mapper
public interface OptionMapper {


    Option findOption(@Param(value = "name") String name);

    String getStringValue(@Param(value = "name") String name);

    Integer getIntegerValue(@Param(value = "name") String name);
}
