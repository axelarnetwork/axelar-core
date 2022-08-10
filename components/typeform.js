import { useEffect } from "react";
import { useSelector, useDispatch, shallowEqual } from "react-redux";
import { Popover } from "@typeform/embed-react";

import { SPENDING_TIME_SECONDS } from "../reducers/types";

export default () => {
  const dispatch = useDispatch();
  const { spending_time } = useSelector(state => ({ spending_time: state.spending_time }), shallowEqual);
  const { spending_time_seconds } = { ...spending_time };

  useEffect(() => {
    if (process.env.NEXT_PUBLIC_TYPEFORM_ID && spending_time_seconds > 0) {
      const interval = setInterval(() => {
        dispatch({
          type: SPENDING_TIME_SECONDS,
          value: spending_time_seconds - 1,
        });
      }, 1000);
      return () => clearInterval(interval);
    }
  }, [spending_time_seconds]);

  return process.env.NEXT_PUBLIC_TYPEFORM_ID && spending_time_seconds < 1 && (
    <Popover
      id={process.env.NEXT_PUBLIC_TYPEFORM_ID}
      customIcon="https://images.typeform.com/images/fVTj4k4SPXfM"
      buttonColor="#FBFBFB"
      notificationDays={7}
      tooltip="Hey 👋&nbsp;&nbsp;How can we help?"
      open="scroll"
      chat={true}
      medium="snippet"
      style={{ all: 'unset' }}
    />
  );
};