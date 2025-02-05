package com.lightstore.user.infrastructure.persistence.repository;

import com.lightstore.user.infrastructure.persistence.mapping.UserModel;
import org.springframework.data.repository.Repository;

public interface JdbcUserRepository extends Repository<UserModel, String> {
    UserModel save(UserModel user);
}
