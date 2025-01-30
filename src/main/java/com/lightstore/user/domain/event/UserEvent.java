package com.lightstore.user.domain.event;

import com.lightstore.user.domain.shared.Event;
import com.lightstore.user.domain.vo.UserID;

public interface UserEvent extends Event {
    UserID getUserID();
}
