// TODO: THIS!
import '@testing-library/jest-dom/vitest'
import {render, screen, within} from '@testing-library/react';
import React from 'react';
import BagDisplay from '@/app/components/bagClient'
import {BagData} from "@/app/components/bagServer";
import {ExamplePicsWithNotes} from "@/app/components/formSubcomponents/picWithNotes";
import {
    ExampleContaminations,
} from "@/app/components/formSubcomponents/contaminations";

describe('View Component', () => {
    it('aTest', () => { // TODO: name
        // TODO: this!
        let bagArea = <BagDisplay readonly={true} data={new BagData({
            _id: "testId",
            recipe: "testRecipe",
            substrateBatch: "batch",
            pcRun: "run",
            filterSize: "size",
            creationDate: 12345,
            genSpore: 3,
            genFruitOrSpore: 2,
            sealDate: 23456,
            wetness: undefined,
            knownFruitable: undefined,
            species: "species",
            subspecies: undefined,
            innoc: "innoc",
            transfersOut: ["1","2","3"],
            parentType: "plate",
            parent: "aPlate",
            pics: ExamplePicsWithNotes,
            contamination: ExampleContaminations,
            mostRecentImage: ExamplePicsWithNotes[1],
            flushes: ExamplePicsWithNotes,
            sale: undefined,
            disposed: undefined,
            notes: ExamplePicsWithNotes[1].notes,
            lastUpdated: 99999,
            acl: {
                users:  (new Map<string, boolean>()).set("testUser", true),
                projects: (new Map<string, boolean>()).set("testProject", true),
                blanketPerm: true,
            },
        })} headerLevel={0} isTopLevel={true} />
        render(bagArea);
        let mainArea = screen.getByRole("main")
        //expect(mainArea).visible()
        let wetness = within(mainArea).getByTestId("wetness-display")
        expect(wetness.innerText === "Wetness: unknown")
        expect(wetness).toHaveRole('status')

        // TODO: expect(comp).toBeInTheDocument()
        // const heading = screen.getByText(/Hello world! I am using React/i);
        // expect(heading).toBeInTheDocument()
    });
});
// describe('Create Component', () => {
//
// });
// describe('Import Component', () => {
//
// });
// TODO: any other components