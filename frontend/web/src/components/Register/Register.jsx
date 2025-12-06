import React, { useState, useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import "./register.css";
import logo from "../../../lib/assets/Kikonect_logo.png";

export default function Register() {
    const location = useLocation();
    const navigate = useNavigate();
    const prefilledEmail = location.state?.email || "";

    const [firstName, setFirstName] = useState("");
    const [lastName, setLastName] = useState("");
    const [email, setEmail] = useState(prefilledEmail);
    const [password, setPassword] = useState("");
    const [confirm, setConfirm] = useState("");

    useEffect(() => {
        if (prefilledEmail) setEmail(prefilledEmail);
    }, [prefilledEmail]);

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (!email || !password || !confirm || !firstName || !lastName) {
            alert("Please fill in all fields.");
            return;
        }
        const emailRegex = /^[\w.-]+@[\w.-]+\.\w+$/;
        if (!emailRegex.test(email)) {
            alert("Please enter a valid email address.");
            return;
        }
        if (password !== confirm) {
            alert("Passwords do not match.");
            return;
        }

        try {
            const res = await fetch("http://localhost:8080/register", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    email,
                    password,
                    firstname: firstName,
                    lastname: lastName,
                }),
            });

            if (res.status === 409) {
                alert("Email already registered. Please login instead.");
                return;
            }

            if (!res.ok) {
                alert("Server error. Please try again later.");
                return;
            }

            const data = await res.json();
            console.log("Registration success:", data);
            alert("Registration successful! You can now login.");
            setEmail("");
            setPassword("");
            setConfirm("");
            setFirstName("");
            setLastName("");
            navigate("/");
        } catch (err) {
            console.error("Network or fetch error:", err);
            alert("Network error. Please check your connection or backend.");
        }
    };

    return (
        <div className="reg-page">
            <div className="reg-card">
                <img src={logo} alt="KiKoNect logo" className="logoR-img" />
                <h2 className="title">Create an account</h2>
                <form onSubmit={handleSubmit} className="reg-form">
                    <div className="floating-input">
                        <input
                            type="text"
                            value={firstName}
                            onChange={(e) => setFirstName(e.target.value)}
                            required
                        />
                        <label className={firstName ? "filled" : ""}>First Name</label>
                    </div>
                    <div className="floating-input">
                        <input
                            type="text"
                            value={lastName}
                            onChange={(e) => setLastName(e.target.value)}
                            required
                        />
                        <label className={lastName ? "filled" : ""}>Last Name</label>
                    </div>
                    <div className="floating-input">
                        <input
                            type="email"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            required
                        />
                        <label className={email ? "filled" : ""}>Email</label>
                    </div>
                    <div className="floating-input">
                        <input
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            required
                        />
                        <label className={password ? "filled" : ""}>Password</label>
                    </div>
                    <div className="floating-input">
                        <input
                            type="password"
                            value={confirm}
                            onChange={(e) => setConfirm(e.target.value)}
                            required
                        />
                        <label className={confirm ? "filled" : ""}>Confirm Password</label>
                    </div>
                    <button type="submit" className="reg-btn">
                        Register
                    </button>
                </form>
            </div>
        </div>
    );
}
