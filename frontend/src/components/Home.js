import React, {useContext, useEffect, useState} from 'react';
import axios from "axios";
import './Login.css';
import {useNavigate} from "react-router-dom";
import LoggedInContext from "../context-providers/loggedin-provider/LoggedInContext";

const Home = () => {
  const navigate = useNavigate();
  const [user, setUser] = useState();
  const {isLoggedIn} = useContext(LoggedInContext)

  useEffect(() => {
    console.log("isLoggedIn", isLoggedIn);
    if (!isLoggedIn) {
      navigate("/login");
    } else {
      axios.get("/api/user-service/v1/user").then(({data}) => {
        console.log(data);
        setUser(data);
      }).catch((err) => {
        console.log(err)
        // navigate("/login");
      });
    }
  }, [isLoggedIn]);

  return (
    <div className="auth-container">
      <p>Hello {user?.email}</p>
    </div>
  );
};

export default Home;
