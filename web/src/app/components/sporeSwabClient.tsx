'use client'

import React, {JSX, useContext, useState} from "react";
import {
    createApiUrlFor,
    DisplayFormWrapper,
    DisplayInput, DoCreateRequest, DoCreateRequestMultipart, DoUpdateRequest, ErrHandler,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    HandleJsonResponse,
    HandleTxtResponse,
    importApiUrlFor,
    ImportDisplayInput,
    ImportEntryFormWrapper,
    IsString,
    ListPageItems,
    ListPageTable,
    ListTableColumn, MultipartImportRequest,
    NewColumn,
    NewEntryFormWrapper,
    NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey,
    SendMultipartRequest,
    setFormData, updateApiUrlFor, viewUrlFor
} from "@/app/components/common";
import {
    DisposedDisplay,
    ErrorDisplay,
    ParentDisplay,
} from "@/app/components/formSubcomponents/commonClient";
import {
    IsValidNote,
    NewEntryNotes,
    Note,
    NoteEntriesGroup, NotesFormArea,
} from "@/app/components/formSubcomponents/notes";
import {SporePrintData} from "@/app/components/sporePrintServer";
import TestAndValidate from "@/app/components/testing/untested";
import {FruitData} from "@/app/components/fruitServer";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {SporeSwabData} from "@/app/components/sporeSwabServer";
import ID from "@/app/components/formSubcomponents/id";
import {ACL} from "@/app/components/accessControlServer";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {BaseExternalUrl} from "@/app/components/Constants";
import {redirect} from "next/navigation";
import {AllEntries, OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import DateArea from "@/app/components/formSubcomponents/date";
import {ExistingSpeciesSelector, SpeciesSubspeciesArea} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {SaleArea} from "@/app/components/saleClient";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {WriteRfidOvcArea} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {OnViewCreatorsTriColArea, OvcForXfers} from "@/app/components/formSubcomponents/ovc";
import {AssertSporePrint} from "@/app/components/sporePrintClient";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {AgarRecipeData} from "@/app/components/agarRecipeServer";

// TODO: list page not working
// TODO: ensure display page doing what we want

export function AssertSporeSwab(input: any): asserts input is SporeSwabData {
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

export function SporeSwabImportDisplay({headerLevel}: ImportDisplayInput) { // TODO: USE ONLY FOR EXISTING SPORE PRINTS!
    const [swabDate, setSwabDate] = useState<number>(Date.now())
    const [notes, setNotes] = useState<Note[]>([])
    const [species, setSpecies] = useState<SpeciesData | undefined>()
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>()
    const [image, setImage] = useState<File | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const cookies = useContext(CookiesContext)
    const importEntry = (e: React.MouseEvent) => {
        e.preventDefault()
        if (!species) {
            setErr("A species must be selected")
            return
        }
        let formData = new FormData()
        let dataObj: any = {
            printDate: swabDate, // TODO: rename from printDate to swabDate???
            species: species._id, // TODO: validate on insert
            subspecies: subspecies?._id, // TODO: validate on insert
            notes: notes,
        }
        setFormData(formData, dataObj)
        formData.set("data", JSON.stringify(dataObj))
        // Img
        if (image) {
            formData.set("img", image, "img")
        }

        MultipartImportRequest(formData, "sporeSwab", AssertSporeSwab, setErr, allCookies(cookies))
        // TODO: reenable if not work: SendMultipartRequest(BaseExternalUrl + "/db/import/sporeSwab", cookies, body)
        // SendMultipartRequest2(importApiUrlFor("sporeSwab"), formData)
        //     .then(HandleJsonResponse)
        //     .then(newItem => {
        //         AssertSporeSwab(newItem)
        //         redirect(viewUrlFor("sporeSwab", newItem._id))
        //     })
        //     .catch(ErrHandler(setErr));
    }
    //no parent because we couldn't possibly know it
    return <ImportEntryFormWrapper entryType={"sporeSwab"}>
        <ErrorDisplay err={err} headerLevel={headerLevel}/>
        <DateArea pre={"Swab Date: "} readonly={false} when={Date.now()} updateParent={setSwabDate}/>
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
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput) {
    try {
        AssertSporeSwab(data)
        const [initial, setInitial] = useState(data)

        const [sale, setSale] = useState(initial.sale)
        const [disposed, setDisposed] = useState(initial.disposed)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: SporeSwabData) => {
            setInitial(updated)
            setSale(updated.sale)
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
        }
        const cookies = useContext(CookiesContext)
        const submit = () => {
            let body: any = {
                sale: sale,
                disposed: disposed,
                notes: notes,
                acl: MarshalAcl(acl),
            }
            DoUpdateRequest("sporeSwab",data._id, body, AssertSporeSwab, allCookies(cookies))
                .then(v=>{
                    updateInitial(new SporeSwabData(v))
                })
                .catch(e=>{
                    setErr(JSON.stringify(e))
                })
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            // TODO: use the next one in other places...
            OvcForXfers(data._id, "sporeSwab", ["plate", "slant", "stasisTube", "jar", "bag", "fruitingChamber"], allCookies(cookies)),
            WriteRfidOvcArea(initial._id),
        ]
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
        return <div>{"ERROR: Spore swab data format incorrect: " + JSON.stringify(err)}</div>
    }
}

// Should only be accessible from a fruit's page
export function NewSporeSwabForm(
    {printIn, fruitIn, headerLevel, offset, onCreate}: {
        printIn?: SporePrintData
        fruitIn?: FruitData
        headerLevel?: number
        offset?: number
        onCreate: (sp: SporeSwabData) => void
    }) {
    // TODO: EITHER PRINT OR FRUIT!!!!!

    // TODO: THIS!
    const [parent, setParent] = useState<string | undefined>()
    const [notes, setNotes] = useState<Note[]>([])

    const [err, setErr] = useState<string | undefined>(undefined)
    const errHandler = ErrHandler(setErr)
    const cookies = useContext(CookiesContext)
    const createEntry = (e: React.MouseEvent) => {
        e.preventDefault()
        if (!parent || parent === "") {
            setErr("Parent must be selected")
            return
        }
        let body: any = {
            parent: parent,
            notes: notes,
        }
        DoCreateRequest("sporeSwab", body, AssertSporeSwab, allCookies(cookies))
            .then(v=>{
                onCreate ? onCreate(v) : console.log("no onCreate provided")
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }

    return <NewEntryFormWrapper entryType={"sporeSwab"}>
        <ErrorDisplay err={err} headerLevel={headerLevel} offset={offset}/>
        {/* TODO: PARENT SELECTOR */}
        <NewEntryNotes setNotes={setNotes}/>
        <button className={"greenButton"} onClick={createEntry}>{"Create"}</button>
    </NewEntryFormWrapper>
}

export function SporeSwabListPageTable({data, onClick, withLink}: ListPageItems<SporeSwabData>) {
    let cols: ListTableColumn<SporeSwabData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Created", (v)=>{
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Spec", (v)=>v.species||""),
        NewColumn("Subspec", v=>v.subspecies||"" ),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: SporeSwabData)=>{
            return <EntryLinkWrapper props={{entry:v,openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new SporeSwabData(v)}}/>
}
export function SporeSwabSelectorTable({data, onClick}: ListPageItems<SporeSwabData>) {
    return <SporeSwabListPageTable data={data} onClick={onClick} withLink={true} />
}

export function SporeSwabSelector(
    {
        doSelect,
        allowCreate // TODO: del?
    }: {
        doSelect: (val: SporeSwabData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: SporeSwabData[]):JSX.Element=>{
        return <SporeSwabSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"sporeSwab"} entryTypes={"sporeSwabs"} doSelect={doSelect} asserter={AssertSporeSwab}
                                   table={table}>
        {/* TODO: ok?allowCreate && <NewSporeSwabForm handlers={{onCreate: doSelect,isTopLevel: false}}/>*/}
    </ExistingRecentSelector>
}
