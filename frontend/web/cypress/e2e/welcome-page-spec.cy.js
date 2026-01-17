import {URL, USER_MAIL, USER_PASSWORD} from "./vars.js";

describe('Navbar links tests', () => {
    it('Logo link', () => {
        cy.visit(URL)

        cy.get('.navbar-logo-link').click()
        cy.url().should('eql', URL + '/')
    })

    it('Login button', () => {
        cy.visit(URL)

        cy.get('.navbar-login-btn > a').click()
        cy.url().should('eql', URL + '/login')
    })

    it('Nav links', () => {
        cy.visit(URL)

        cy.get('.navbar-list > :nth-child(1) > a').click()
        cy.url().should('eql', URL + '/#features')

        cy.get('.navbar-list > :nth-child(2) > a').click()
        cy.url().should('eql', URL + '/#how-it-works')
    })
});

describe('Content links tests', () => {
    it('Get started button', () => {
        cy.visit(URL)

        cy.get('.welcome-page-section-pres-btns > .how-cta-btn').click()
        cy.url().should('eql', URL + '/login')
    })
});

describe('Footer links tests', () => {
    it('Footer nav links', () => {
        cy.visit(URL)

        cy.get(':nth-child(2) > :nth-child(2) > a').click()
        cy.url().should('eql', URL + '/#features')

        cy.get(':nth-child(2) > :nth-child(3) > a').click()
        cy.url().should('eql', URL + '/#how-it-works')
    })
});

describe('Action flows', () => {
    it('Access to App (not logged)', () => {
        cy.visit(URL + "/home")

        cy.url().should('eql', URL + '/')
    })

    it('Register and login', () => {
        cy.visit(URL + "/createacc")
        cy.get('.reg-card').should('be.visible')

        cy.get(':nth-child(1) > input').click()
        cy.get(':nth-child(1) > input').type(USER_MAIL);

        cy.get('#root button.reg-btn').click()
        cy.get('.reg-card').should('be.visible')
        cy.url().should('eql', URL + '/register')

        cy.get(':nth-child(1) > input').click()
        cy.get(':nth-child(1) > input').type('Test')
        cy.get(':nth-child(2) > input').click()
        cy.get(':nth-child(2) > input').type('User')
        cy.get(':nth-child(4) > input').click()
        cy.get(':nth-child(4) > input').type(USER_PASSWORD)
        cy.get(':nth-child(5) > input').click()
        cy.get(':nth-child(5) > input').type(USER_PASSWORD)
        cy.get('#root button.reg-btn').click();
        cy.get('.login-card').should('be.visible')
        cy.url().should('eql', URL + '/login')

        cy.get(':nth-child(1) > input').click()
        cy.get(':nth-child(1) > input').type(USER_MAIL)
        cy.get(':nth-child(2) > input').click()
        cy.get(':nth-child(2) > input').type(USER_PASSWORD)
        cy.get('[type="submit"]').click()
        cy.url().should('eql', URL + '/home')
    })

    it('Access to App (logged)', () => {
        cy.visit(URL + "/home")

        cy.url().should('eql', URL + '/home')

        cy.get('.navbar-login-btn > a').click()
    })
});