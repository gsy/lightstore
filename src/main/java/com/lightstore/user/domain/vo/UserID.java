package com.lightstore.user.domain.vo;

import com.lightstore.user.domain.shared.RandomUUID;

public class UserID extends RandomUUID {
    public UserID() {
            super();
    }

    public UserID(String id) {
        super(id);
    }

    @Override
    protected String getPrefix() {
        return "%s";
    }
}
