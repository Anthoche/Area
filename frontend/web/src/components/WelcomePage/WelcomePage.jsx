"use client";

import { useEffect, useState, useRef } from "react";
import "./welcomepage.css";
import "./hero-animations.css";
import Navbar from "./Navbar.jsx";
import Footer from "./Footer.jsx";
import HowItWorks from "./HowItWorks.jsx";
import PopUseCases from "./PopUseCases.jsx";
import WhyUs from "./WhyUs.jsx";
import { Link } from "react-router-dom";

export default function WelcomePage() {
    const heroRef = useRef(null);
    const heroIconRef = useRef(null);
    const heroStatsRef = useRef(null);
    const [heroVisible, setHeroVisible] = useState(false);
    const [iconVisible, setIconVisible] = useState(false);
    const [statsVisible, setStatsVisible] = useState(false);

    useEffect(() => {
        setTimeout(() => setHeroVisible(true), 100);
        setTimeout(() => setIconVisible(true), 400);
        setTimeout(() => setStatsVisible(true), 700);
    }, []);
    const userId = Number(localStorage.getItem("user_id"));
    const isLoggedIn = Number.isFinite(userId) && userId > 0;
    const [stats, setStats] = useState({
        userCount: null,
        serviceCount: null,
        actionReactionCount: null,
    });

    const API_BASE =
        import.meta.env.VITE_API_URL ||
        import.meta.env.API_URL ||
        `${window.location.protocol}//${window.location.hostname}:8080`;

    useEffect(() => {
        let cancelled = false;
        const fetchStats = async () => {
            try {
                const res = await fetch(`${API_BASE}/areas`);
                if (!res.ok) throw new Error("failed to load areas");
                const data = await res.json();
                const services = Array.isArray(data.services) ? data.services : [];
                const actionReactionCount = services.reduce(
                    (sum, service) =>
                        sum +
                        (service?.triggers?.length || 0) +
                        (service?.reactions?.length || 0),
                    0
                );
                const visibleServices = services.filter((service) => !service?.hidden);
                if (!cancelled) {
                    setStats({
                        userCount: Number.isFinite(data.user_count)
                            ? data.user_count
                            : 0,
                        serviceCount: visibleServices.length,
                        actionReactionCount,
                    });
                }
            } catch (err) {
                console.error(err);
            }
        };
        fetchStats();
        return () => {
            cancelled = true;
        };
    }, [API_BASE]);

    return (
        <div className="welcome-page-wrapper">
            <div className="welcome-wave-bg">
                <svg viewBox="0 0 1440 220" fill="none" 
                    xmlns="http://www.w3.org/2000/svg" 
                    width="100%" height="220" preserveAspectRatio="none" style={{display: 'block'}}>
                    <defs>
                        <linearGradient id="waveGradient" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="0%" stopColor="#b3e0ff"/>
                            <stop offset="100%" stopColor="#e3f0ff"/>
                        </linearGradient>
                    </defs>
                    <path d="M0,120 C360,200 1080,40 1440,120 L1440,0 L0,0 Z" fill="url(#waveGradient)" fillOpacity="0.4"/>
                    <path d="M0,180 C400,100 1040,260 1440,180 L1440,0 L0,0 Z" fill="#e3f0ff" fillOpacity="0.2"/>
                </svg>
            </div>
            <Navbar />
            <div className="welcome-page-content">
                <div className="welcome-page-section" id="features">
                    <div
                        className={`welcome-page-section-pres hero-animate${heroVisible ? ' visible' : ''}`}
                        ref={heroRef}
                    >
                        <div
                            className={`welcome-page-section-pres-icon hero-shape-animate${iconVisible ? ' visible' : ''}`}
                            ref={heroIconRef}
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                                 stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                                 className="lucide lucide-sparkles" aria-hidden="true">
                                <path
                                    d="M11.017 2.814a1 1 0 0 1 1.966 0l1.051 5.558a2 2 0 0 0 1.594 1.594l5.558 1.051a1 1 0 0 1 0 1.966l-5.558 1.051a2 2 0 0 0-1.594 1.594l-1.051 5.558a1 1 0 0 1-1.966 0l-1.051-5.558a2 2 0 0 0-1.594-1.594l-5.558-1.051a1 1 0 0 1 0-1.966l5.558-1.051a2 2 0 0 0 1.594-1.594z"></path>
                                <path d="M20 2v4"></path>
                                <path d="M22 4h-4"></path>
                                <circle cx="4" cy="20" r="2"></circle>
                            </svg>
                            <span>Connect your apps and automate workflows</span>
                        </div>
                        <h1 className={`hero-animate${heroVisible ? ' visible' : ''}`}>Make your apps work together seamlessly</h1>
                        <p className={`hero-animate${heroVisible ? ' visible' : ''}`}
                            style={{ transitionDelay: '0.2s' }}
                        >
                            Connect your favorite apps and services to automate workflows. No code required.
                            Create powerful automations in minutes and boost your productivity.
                        </p>
                        <div className={`welcome-page-section-pres-btns hero-animate${heroVisible ? ' visible' : ''}`}
                            style={{ transitionDelay: '0.3s' }}
                        >
                            <Link
                                className="how-cta-btn pres-btn-get-started pres-btn"
                                to={isLoggedIn ? "/home" : "/login"}
                            >
                                <span>Get Started</span>
                                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                                     stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                                     className="lucide lucide-arrow-right group-hover:translate-x-1 transition-transform"
                                     aria-hidden="true">
                                    <path d="M5 12h14"></path>
                                    <path d="m12 5 7 7-7 7"></path>
                                </svg>
                            </Link>
                        </div>
                    </div>
                    <div
                        className={`welcome-page-section-stats hero-shape-animate-right${statsVisible ? ' visible' : ''}`}
                        ref={heroStatsRef}
                        style={{transitionDelay: statsVisible ? '0.5s' : '0s' }}
                    >
                        <ul>
                            <li>
                                <h1>{stats.userCount ?? "…"}</h1>
                                <span>Active users</span>
                            </li>
                            <li>
                                <h1>{stats.serviceCount ?? "…"}</h1>
                                <span>Available services</span>
                            </li>
                            <li>
                                <h1>{stats.actionReactionCount ?? "…"}</h1>
                                <span>Actions/Reactions</span>
                            </li>
                        </ul>
                    </div>
                </div>
                <WhyUs />
                <PopUseCases />
                <HowItWorks />
            </div>
            <Footer />
        </div>
    );
}
