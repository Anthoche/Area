"use client";
import "./welcomepage.css";
import "./hero-animations.css";
import { useEffect, useRef, useState } from "react";

export default function WhyUs() {
    const sectionRef = useRef(null);
    const [visible, setVisible] = useState(false);
    useEffect(() => {
        const observer = new window.IntersectionObserver(
            ([entry]) => setVisible(entry.isIntersecting),
            { threshold: 0.2 }
        );
        if (sectionRef.current) observer.observe(sectionRef.current);
        return () => observer.disconnect();
    }, []);
    return (
        <div ref={sectionRef} className={`welcome-page-section why-us hero-animate${visible ? ' visible' : ''}`} id="why-us">
            <h2 className="why-title">Why users trust us</h2>
            <p className="why-subtitle">The most powerful automation platform that helps you work smarter, not harder</p>
            <div className="why-steps">
                <div className={`why-step hero-shape-animate${visible ? ' visible' : ''}`} style={{ transitionDelay: visible ? '0.1s' : '0s' }}>
                    <div className="why-emoji" style={{ background: "linear-gradient(135deg, #ffffff4c, #ecf244ff)"}}>⚡️</div>
                    <h3>Lightning fast automation</h3>
                    <p>Set up powerful automations in minutes without writing a single line of code. Our intuitive interface makes it simple.</p>
                </div>
                <div className={`why-step hero-shape-animate${visible ? ' visible' : ''}`} style={{ transitionDelay: visible ? '0.25s' : '0s' }}>
                    <div className="why-emoji" style={{ background: "linear-gradient(135deg, white, #40c0ffff)" }}>🔗</div>
                    <h3>Connect anything</h3>
                    <p>Choose from 800+ apps and services. From Gmail to Slack, from Twitter to Google Sheets - we support them all.</p>
                </div>
                <div className={`why-step hero-shape-animate${visible ? ' visible' : ''}`} style={{ transitionDelay: visible ? '0.4s' : '0s' }}>
                    <div className="why-emoji" style={{ background: "linear-gradient(135deg, white, #58ba32ff)" }}>✔</div>
                    <h3>Always reliable</h3>
                    <p>Your automations run 24/7 in the cloud. Set it and forget it. We handle the infrastructure so you don't have to.</p>
                </div>  
            </div>
        </div>
    );
}

