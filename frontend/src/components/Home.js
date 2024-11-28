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
    if (!isLoggedIn) {
      console.log("navigating to login", isLoggedIn)
      navigate("/login");
    }
  }, [isLoggedIn]);

  useEffect(() => {
    axios.get("/api/user-service/v1/user").then(({data}) => {
      console.log(data);
      setUser(data);
    }).catch((err) => {
      console.log(err)
      // navigate("/login");
    });
  }, []);

  return (
    <div className="auth-container">
      <p>Hello {user?.email}</p>
    </div>
  );
};

export default Home;
