import {URL, USER_MAIL, USER_PASSWORD} from "./vars.js";

describe('App tests', () => {
    it('Login', () => {
        cy.visit(URL + "/login")
        cy.get(':nth-child(1) > input').click()
        cy.get(':nth-child(1) > input').type(USER_MAIL)
        cy.get(':nth-child(2) > input').click()
        cy.get(':nth-child(2) > input').type(USER_PASSWORD)
        cy.get('[type="submit"]').click()
        cy.url().should('eql', URL + '/home')
    })
});

describe('Action flows', () => {
    it('Create konect', () => {
        cy.visit(URL + "/home")

        cy.url().should('eql', URL + '/home')
    })
})