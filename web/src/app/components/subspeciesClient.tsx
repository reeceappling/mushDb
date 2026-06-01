'use client'

import React, {JSX, useContext, useEffect, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {AllEntries} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {
    AssertArrayResult, clientPostRequestHeaders, createApiUrlFor,
    CreateNewEntryButton, DisplayFormWrapper,
    DisplayInput, DoCreateRequest, DoUpdateRequest, ErrHandler, ExistingRecentSelector, FlexedArea, FlexedSinglesGroup,
    HandleJsonResponse,
    IsString, ListPageItems, ListPageTable, ListTableColumn, NewColumn, NewEntryFormWrapper,
    NewEntryInput, NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey, Subform, updateApiUrlFor,
} from "@/app/components/common";
import {AliasesArea, ErrorDisplay, NameArea} from "@/app/components/formSubcomponents/commonClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {AssertSpecies, ExistingSpeciesSelector} from "@/app/components/speciesClient";
import {
    AclDefaultAclDisplay,
    IsValidAcl, MarshalAcl
} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import TestAndValidate from "@/app/components/testing/untested";
import {HandleErr} from "@/app/components/userClient";
import {SpeciesData} from "@/app/components/speciesServer";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {AssertStasisTube} from "@/app/components/stasisTubeClient";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";

export function AssertSubspecies(input: any): asserts input is SubspeciesData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['species', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Subspecies assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['acl', IsValidAcl],
        ['defaultAcl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Subspecies assertion failure: optional key ' + key + ' was not valid: ');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
        ['aliases', IsString],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Subspecies assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export default function SubspeciesDisplay(
    {
        id, readonly, data, headerLevel
    }: DisplayInput) {
    try {
        AssertSubspecies(data)
        const [initial, setInitial] = useState(data)

        const [aliases, setAliases] = useState(data.aliases || [])
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const [defaultAcl, setDefaultAcl] = useState<ACL | undefined>(initial.defaultAcl)
        const updateInitial = (updated: SubspeciesData) => {
            setInitial(updated)
            setAliases(updated.aliases || [])
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
            setDefaultAcl(updated.defaultAcl)
        }
        const cookies = useContext(CookiesContext)
        const update = () => {
            const body: any = {
                aliases: aliases,
                notes: notes,
                acl: MarshalAcl(acl), // TODO: ensure ok
                defaultAcl: MarshalAcl(defaultAcl), // TODO: ensure ok
            }
            DoUpdateRequest("subspecies",encodeURIComponent(initial._id), body, AssertSubspecies, allCookies(cookies))
                .then(updateInitial)
                .catch(ErrHandler(setErr))
            // fetch(updateApiUrlFor("subspecies",encodeURI(initial._id)), {
            //     method: "POST",
            //     headers: clientPostRequestHeaders,
            //     body: JSON.stringify(body)
            // })
            //     .then(HandleJsonResponse)
            //     .then((entry) => {
            //         AssertSubspecies(entry)
            //         updateInitial(entry)
            //     })
            //     .catch((error) => {
            //         HandleErr(error, setErr)
            //     });
        }
        return (
            <DisplayFormWrapper entryType={"subspecies"}>
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                <TestAndValidate todos={["Species up here too?"]}>
                    <ID id={data._id} txt={"Subspecies"} entryType={"subspecies"} />
                </TestAndValidate>
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                    </FlexedSinglesGroup>
                </FlexedArea>
                <AliasesArea aliases={aliases} readonly={readonly} headerLevel={headerLevel} updateParent={setAliases}/>
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <AclDefaultAclDisplay ACL={acl} defaultACL={defaultAcl} updateAcl={setAcl} updateDefaultAcl={setDefaultAcl} readonly={readonly}/>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                    e.stopPropagation();
                    update()
                }}>{"Update Subspecies"}</button>}
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: Subspecies data format incorrect: " + err}</div>
    }

}

export function NewSubspeciesForm({handlers, species}: {
    handlers: NewEntryInput<SubspeciesData>,
    species?: SpeciesData
}) {
    const [name, setName] = useState<string | undefined>()
    const [selectedSpecies, setSelectedSpecies] = useState(species)
    const [aliases, setAliases] = useState<string[]>([])
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>(undefined)
    const errHandler = ErrHandler(setErr)
    const cookies = useContext(CookiesContext)
    const submitNewSubspecies = () => {
        if (!name) {
            setErr("Name must note be blank")
            return
        }
        if (!selectedSpecies) {
            setErr("Species must be selected")
            return
        }
        const body: any = {
                name: name,
                species: selectedSpecies,
                aliases: aliases,
                notes: notes,
            }
        DoCreateRequest("subspecies", body, AssertSubspecies, allCookies(cookies))
            .then(handlers?.onCreate)
            .catch(errHandler)
        // fetch(createApiUrlFor("subspecies"), {
        //     method: "POST",
        //     headers: clientPostRequestHeaders,
        //     body: JSON.stringify(body)
        // })
        //     .then(HandleJsonResponse)
        //     .then((entry) => {
        //         AssertSubspecies(entry)
        //         onCreate && onCreate(entry)
        //     })
        //     .catch(ErrHandler(setErr));
    }
    return (
        <NewEntryFormWrapper entryType={"subspecies"}>
            <ErrorDisplay err={err}/>
            {species === undefined && <ExistingSpeciesSelector initialSpecies={species} doSelect={s => {
                setSelectedSpecies(s)
            }} />}
            {/* NAME (ID) */}
            <NameArea classNames={"inlineChildren"} currentName={name} headerTxt={"New Subspecies Name: "} setName={setName} readonly={false}/>
            {/* Aliases */}
            <AliasesArea aliases={aliases} readonly={false} updateParent={setAliases}/>
            {/* Notes */}
            <NewEntryNotes setNotes={setNotes}/>
            <CreateNewEntryButton onSubmit={submitNewSubspecies}/>
        </NewEntryFormWrapper>
    )
}

// ExistingSubSpeciesSelector selects between subspecies of a SINGLE species!
export function ExistingSubSpeciesSelector(
    {
        species,
        doSelect,
        headerLevel,
    }: {
        species?: string,
        doSelect: (val: SubspeciesData | undefined) => void,
        headerLevel?: number
    }) {
    const cookies = useContext(CookiesContext)
    const [isLoaded, setLoaded] = useState(false)
    const [selectable, setSelectable] = useState(false)
    const [selectorOpen, setSelectorOpen] = useState(false)
    const [subspeciesList, setSubspeciesList] = useState<SubspeciesData[]>([]);
    const [selected, setSelected] = useState<SubspeciesData | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    useEffect(() => {
        if (!species){
            return
        }
        setSelected(undefined)
        setLoaded(false)
        fetch(BaseExternalUrl + "/subspeciesFor/" + encodeURI(species), { // TODO: ensure endpoint ok
            method: "GET",
            headers: clientPostRequestHeaders,
        })
            .then(HandleJsonResponse)
            .then((data) => {
                AssertArrayResult(data, AssertSubspecies)
                setSubspeciesList(data as SubspeciesData[])
                setLoaded(true)
                setSelectable(species !== undefined)
                setErr(undefined)
            })
            .catch(ErrHandler(setErr));
    }, [species]);
    let errArea = () => {
        return <ErrorDisplay err={err} headerLevel={headerLevel}/>
    }
    const toggleButton = () => {
        return <div>
            <button className={"basicButton"} onClick={() => {
                setSelectorOpen(!selectorOpen)
            }}>{selectorOpen ? "Close subspecies selector" : (selected ? "Choose a different subspecies" : "Select a subspecies")}</button>
        </div>
    }
    if (!selectable) {
        return null
    }
    if (!selected && !selectorOpen) {
        return <div className={"centerHChildren gapTop gapBottom"}>
            {errArea()}
            {toggleButton()}
        </div>
    }
    if (!isLoaded) {
        return <div className={"centerHChildren gapTop gapBottom"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <div>{"loading subspecies selector"}</div>
        </div>
    }
    if (subspeciesList.length == 0) {
        return <div className={"centerHChildren gapTop gapBottom"}>
            <ErrorDisplay err={"No Subspecies Found for species: " + (species && species)} headerLevel={headerLevel}/>
            <TestAndValidate todos={["do this"]}>
                <div>{"CREATE SUBSPECIES LINK"}</div>
            </TestAndValidate>
        </div>
    }
    if (selected && !selectorOpen) {
        return <div className={"centerHChildren gapTop gapBottom"}>
            {errArea()}
            {"Currently Selected subspecies: " + selected._id}
            {toggleButton()}
        </div>
    }

    return <div className={"centerHChildren gapTop gapBottom"}>
        <Subform>
        {errArea()}
        {toggleButton()}
        <SubspeciesSelector doSelect={s=>{
            setSelected(s)
            setSelectorOpen(false)
            doSelect(s)
        }} />
        {toggleButton()}
        </Subform>
    </div>
}

export function SubspeciesFormArea({subspecies}:{
    subspecies: string,
}){
    return <EntryLinkWrapper props={{entryType:"subspecies",linkId:encodeURI(subspecies)}}><div>{"Subspecies: "+subspecies}</div></EntryLinkWrapper>
}

export function SubspeciesListPageTable({data, onClick, withLink}: ListPageItems<SubspeciesData>) {
    let cols: ListTableColumn<SubspeciesData>[] = [
        NewColumn("Subspecies", (v)=>v._id),
        NewColumn("Species", (v)=>v.species),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: SubspeciesData)=>{
            return <EntryLinkWrapper props={{linkId:encodeURI(v._id),entryType:"subspecies",openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}
export function SubspeciesSelectorTable({data, onClick}: ListPageItems<SubspeciesData>) {
    return <SubspeciesListPageTable data={data} onClick={onClick} withLink={true} />
}

// SubspeciesSelector is a selector between ALL subspecies, not just those of a single species
export function SubspeciesSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: SubspeciesData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: SubspeciesData[]):JSX.Element=>{
        return <SubspeciesSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"subspecies"} entryTypes={"subspecies"} doSelect={doSelect} asserter={AssertSubspecies}
                                   table={table}>
    </ExistingRecentSelector>
}