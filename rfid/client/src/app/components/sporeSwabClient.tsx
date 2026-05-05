'use client'

import React, {useState} from "react";
import {
    DisplayInput, HandleJsonResponse, HandleTxtResponse,
    ImportDisplayInput, InlineExpansionArea, InlineExpansionButton,
    InlineProps, InlineSubArea,
    IsString,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey,
    SendMultipartRequest, setFormData
} from "@/app/components/common";
import {
    DisposedDisplay,
    ErrorDisplay,
    ParentDisplay,
    SpeciesArea, SubspeciesArea,
} from "@/app/components/formSubcomponents/commonClient";
import {
    IsValidNote,
    NewEntryNotes,
    Note,
    NoteEntriesGroup,
    NotesAreaInline
} from "@/app/components/formSubcomponents/notes";
import {SporePrintData} from "@/app/components/sporePrintServer";
import TestAndValidate from "@/app/components/testing/untested";
import {FruitData} from "@/app/components/fruitServer";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {SporeSwab} from "@/app/components/sporeSwabServer";
import {DisplayFormWrapper, ImportEntryFormWrapper, NewEntryFormWrapper} from "@/app/components/lcRecipeClient";
import {InlineEntry} from "@/app/components/agarRecipeClient";
import ID from "@/app/components/formSubcomponents/id";
import {FlexedArea, FlexedSinglesGroup, NotesFormArea} from "@/app/components/agarBatchClient";
import {ACL} from "@/app/components/accessControlServer";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {BaseExternalUrl} from "@/app/components/Constants";
import {redirect} from "next/navigation";
import {AllEntries, OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import DateArea from "@/app/components/formSubcomponents/date";
import {ExistingSpeciesSelector} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {SaleArea} from "@/app/components/saleClient";
import {OvcForXfers} from "@/app/components/bagClient";
import {OnViewCreatorsTriColArea} from "@/app/components/pcRunClient";
import {SpeciesSubspeciesArea} from "@/app/components/lcClient";

// TODO: list page not working
// TODO: ensure display page doing what we want

export function AssertSporeSwab(input: any): asserts input is SporeSwab {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // TODO: THIS!

    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['species', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Plate assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['parent', 'string'],
        ['parentType', 'string'],
        ['subspecies', 'string'],
        ['sale', 'string'],
        ['disposed', 'number'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Swab assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Swab assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', IsString],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Swab assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export function SporeSwabImportDisplay({headerLevel, cookies}: ImportDisplayInput) { // TODO: USE ONLY FOR EXISTING SPORE PRINTS!
    const [printDate, setPrintDate] = useState<number>(Date.now())
    const [notes, setNotes] = useState<Note[]>([])
    const [species, setSpecies] = useState<SpeciesData | undefined>()
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>()
    const [image, setImage] = useState<File | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const importEntry = (e: React.MouseEvent) => {
        e.preventDefault()
        if (!species) {
            setErr("A species must be selected")
            return
        }
        let body = new FormData()
        let dataObj: any = {
            printDate: printDate,
            species: species._id,
            subspecies: subspecies?._id, // TODO: validate on insert
            notes: notes,
        }
        setFormData(body, dataObj)
        body.set("data", JSON.stringify(dataObj))
        // Img
        if (image) {
            body.set("img", image, "img") // TODO: ensure works
        }

        SendMultipartRequest(BaseExternalUrl + "/db/import/sporePrint", cookies, body)
            .then(HandleTxtResponse)
            .then(id => { // TODO: ensure txt response is ok here
                redirect(BaseExternalUrl + "/view/sporePrint/" + id)
            })
            .catch((er) => {
                setErr(JSON.stringify(er))
            });
    }
    //no parent because we couldn't possibly know it
    return <ImportEntryFormWrapper entryType={"sporeSwab"}>
        <ErrorDisplay err={err} headerLevel={headerLevel}/>
        <DateArea pre={"Print Date: "} readonly={false} when={Date.now()} updateParent={setPrintDate}/>
        <ExistingSpeciesSelector doSelect={setSpecies} headerLevel={headerLevel}/>
        <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies} headerLevel={headerLevel}/>
        <ImageSelector updateParent={setImage}/>
        <NoteEntriesGroup preexisting={false} readonly={false} updateParent={ns => {
            setNotes(ns.map(n => {
                return n.data
            }))
        }} headerLevel={headerLevel}/>
        <button className={"greenButton"} onClick={importEntry}>{"Create"}</button>
    </ImportEntryFormWrapper>

}

export default function SporeSwabDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel, cookies,
    }: DisplayInput) {
    try {
        AssertSporeSwab(data)
        const [initial, setInitial] = useState(data)

        const [sale, setSale] = useState(initial.sale)
        const [disposed, setDisposed] = useState(initial.disposed)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: SporeSwab) => {
            setInitial(updated)
            setSale(updated.sale)
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
        }
        const submit = () => {
            let dataObj: any = {
                sale: sale,
                disposed: disposed,
                notes: notes,
                acl: MarshalAcl(acl),
            }

            fetch(BaseExternalUrl + "/db/update/sporeSwab/" + data._id, {
                method: 'Post',
                body: JSON.stringify(dataObj),
                headers: {
                    credentials: 'include',
                    'Content-type': "application/json"
                },
            })
                .then(HandleJsonResponse)
                .then((entry) => {
                    AssertSporeSwab(entry)
                    updateInitial(entry)
                })
                .catch((er) => {
                    setErr(JSON.stringify(er))
                });
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            OvcForXfers(data._id, "sporeSwab", ["plate", "slant", "stasisTube", "jar", "bag", "fruitingChamber"], cookies), // TODO: ensure list correct???
        ] // TODO: THIS!
        return <DisplayFormWrapper entryType={"sporeSwab"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <ID id={id} entryType={"sporeSwab"} txt={"Spore Swab"}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <TestAndValidate todos={["reformat these into groups"]}>
                        <ParentDisplay parent={initial.parent} parentType={initial.parentType} />
                        <SaleArea readonly={false} canCreateSale={true} sale={sale} setSale={setSale}
                                  headerLevel={headerLevel}/>
                        <DateArea pre={"Print Date: "} readonly={true} when={initial.creationDate}/>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                        <DisposedDisplay readonly={false} disposed={disposed} setDisposedOnParent={setDisposed}/>
                        <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                        {/*<SpeciesSubspeciesFormArea species={initial.species} subspecies={initial.subspecies}/>*/}
                       </TestAndValidate>
                </FlexedSinglesGroup>
            </FlexedArea>

            <NotesFormArea initial={initial.notes} readonly={readonly} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl} />
            </TogglableAreaWithDepth>
            {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                e.stopPropagation();
                submit()
            }}>{"Update"}</button>}
            <OnViewCreatorsTriColArea OnViewCreators={ovcs}
                                      readonly={readonly}/> {/*swab to agar and that's about it */}
        </DisplayFormWrapper>
    } catch (err) {
        return <div>{"ERROR: Spore swab data format incorrect: " + err}</div>
    }// TODO: VALIDATE WORKS AS EXPECTED
}

// Should only be accessible from a fruit's page
export function NewSporeSwabForm(
    {printIn, fruitIn, headerLevel, offset, onCreate}: {
        printIn?: SporePrintData
        fruitIn?: FruitData
        headerLevel?: number
        offset?: number
        onCreate: (sp: SporeSwab) => void
    }) {
    // TODO: EITHER PRINT OR FRUIT!!!!!

    // TODO: THIS!
    const [parent, setParent] = useState<string | undefined>()
    const [notes, setNotes] = useState<Note[]>([])

    const [err, setErr] = useState<string | undefined>(undefined)
    const createEntry = (e: React.MouseEvent) => {
        e.preventDefault()
        if (!parent || parent === "") {
            setErr("Parent must be selected")
            return
        }
        let dataObj: any = {
            parent: parent,
            notes: notes,
        }

        fetch(BaseExternalUrl + "/db/create/sporeSwab", {
            method: 'Post',
            body: JSON.stringify(dataObj),
            headers: {
                credentials: 'include',
                'Content-type': "application/json"
            },
        })
            .then(HandleJsonResponse)
            .then((resJson) => {
                AssertSporeSwab(resJson) // TODO: make sure comes back as swab obj?
                onCreate(resJson)
            })
            .catch((er) => {
                setErr(JSON.stringify(er))
            });
    }

    return <NewEntryFormWrapper entryType={"sporeSwab"}>
        <ErrorDisplay err={err} headerLevel={headerLevel} offset={offset}/>
        {/* TODO: PARENT SELECTOR */}
        <NewEntryNotes setNotes={setNotes}/>
        <button className={"greenButton"} onClick={createEntry}>{"Create"}</button>
    </NewEntryFormWrapper>// TODO: VALIDATE WORKS AS EXPECTED
}

export function SporeSwabInline(
    {
        data, expandByDefault, headerLevel, onClick, showMainPageButton, idIsLink
    }: InlineProps<SporeSwab>
) {
    const [expanded, setExpanded] = useState(expandByDefault)
    return <InlineEntry onClick={onClick}>
        <InlineSubArea props={{}}>
            <ID id={data._id} txt={"Spore Swab"} entryType={"sporeSwab"} allowOpenMainPage={showMainPageButton}
                linkPage={idIsLink}/>
            <DateArea pre={"Creation Date: "} readonly={true} when={data.creationDate}/>
            <SpeciesArea readonly={true} headerLevel={headerLevel} initial={data.species}/>
            <SubspeciesArea readonly={true} headerLevel={headerLevel} currentSpecies={data.species}
                            initialSub={data.subspecies}/>
        </InlineSubArea>
        <InlineExpansionArea props={{expanded: expanded}}>
            <div>
                <ParentDisplay parent={data.parent} parentType={data.parentType}/>
            </div>
            <NotesAreaInline notes={data.notes} offset={-1}/>
            <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
        </InlineExpansionArea><InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded} expanded={expanded}/>

    </InlineEntry>// TODO: VALIDATE WORKS AS EXPECTED
}
