'use client'

import React, {JSX, useContext, useEffect, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {AllEntries} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {
    AssertArrayResult, clientPostRequestHeaders,
    CreateNewEntryButton, DisplayFormWrapper,
    DisplayInput, DoCreateRequest, DoUpdateRequest, ErrHandler, ExistingRecentSelector, FlexedArea, FlexedSinglesGroup,
    HandleJsonResponse,
    IsString, ListPageItems, ListPageTable, ListTableColumn, NewColumn, NewEntryFormWrapper,
    NewEntryInput, NumberToDateStr,
    OptionalArrayOfType, Subform,
} from "@/app/components/common";
import {AliasesArea, ErrorDisplay, NameArea} from "@/app/components/formSubcomponents/commonClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import { ExistingSpeciesSelector} from "@/app/components/speciesClient";
import {
    AclDefaultAclDisplay,
    MarshalAcl, UnmarshalAcl
} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import TestAndValidate from "@/app/components/testing/untested";
import {SpeciesData, SpeciesSelectorCloseable} from "@/app/components/speciesServer";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";

export function AssertSubspecies(input: any): asserts input is SubspeciesData {
    if (typeof input !== 'object') {
        throw 'Input is not an object! Input is ' + typeof input
    }
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['species', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw 'Subspecies assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key])
        }
    }
    // // complex required keys
    // const complexRequiredKeys = new Map<string, (v: any) => boolean>([
    //     // ['acl', IsValidAcl],
    //     // ['defaultAcl', IsValidAcl]
    // ])
    // for (const [key, validator] of complexRequiredKeys) {
    //     if (!RequiredKey(key, input, validator)) {
    //         throw 'Subspecies assertion failure: required key ' + key + ' was not valid: '
    //     }
    // }
    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
        ['aliases', IsString],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw 'Subspecies assertion failure: optional array key ' + key + ' was not valid'
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    // Unmarshal Default ACL
    if (!('defaultAcl' in input)) {
        throw 'Default Acl missing from input in asserter'
    }
    input.defaultAcl = UnmarshalAcl(input.defaultAcl)
    return
}

export default function SubspeciesDisplay(
    {
        id, readonly, data, headerLevel
    }: DisplayInput<SubspeciesData>) {
        const [initial, setInitial] = useState(data)

        const [aliases, setAliases] = useState(data.aliases || [])
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL>(initial.acl)
        const [defaultAcl, setDefaultAcl] = useState<ACL>(initial.defaultAcl)
        const updateInitial = (updated: SubspeciesData) => {
            setInitial(updated)
            setAliases(updated.aliases || [])
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
            setDefaultAcl(updated.defaultAcl)
            setErr(undefined)
        }
        const cookies = useContext(CookiesContext)
        const update = () => {
            const body: any = {
                aliases: aliases,
                notes: notes,
                acl: MarshalAcl(acl),
                defaultAcl: MarshalAcl(defaultAcl),
            }
            DoUpdateRequest("subspecies",encodeURIComponent(initial._id), body, AssertSubspecies, allCookies(cookies))
                .then(v=>{
                    updateInitial(new SubspeciesData(v))
                })
                .catch(e=>{
                    setErr("failed to update initial: "+JSON.stringify(e))
                })
        }
        return (
            <DisplayFormWrapper entryType={"subspecies"}>
                <ErrorDisplay err={err}/>
                <ID props={{id:data._id, txt:"Subspecies", entryType:"subspecies"}}/>
                <ID props={{id:data.species, txt:"Species", entryType:"species"}}/> {/* TODO: link not working!*/}
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                    </FlexedSinglesGroup>
                </FlexedArea>
                <AliasesArea initial={initial.aliases} readonly={readonly} updateParent={setAliases}/>
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <AclDefaultAclDisplay ACL={acl} defaultACL={defaultAcl} updateAcl={setAcl} updateDefaultAcl={setDefaultAcl} readonly={readonly}/>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                    e.stopPropagation();
                    update()
                }}>{"Update Subspecies"}</button>}
            </DisplayFormWrapper>
        )
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
                species: selectedSpecies._id,
                aliases: aliases,
                notes: notes,
                // ACL/DefaultACL are initially inherited from parent species
            }
        DoCreateRequest("subspecies", body, AssertSubspecies, allCookies(cookies))
            .then(v=>{
                handlers.onCreate ? handlers.onCreate(new SubspeciesData(v)) : console.log("no onCreate provided")
            })
            .catch(e=>{
                console.error("onCreate failed: "+JSON.stringify(e)) // TODO: del
                setErr("onCreate failed: "+JSON.stringify(e))
            })
    }
    return (
        <NewEntryFormWrapper entryType={"subspecies"}>
            <ErrorDisplay err={err}/>
            {/*{species === undefined && <SpeciesSelectorCloseable doSelect={setSelectedSpecies} allowCreation={false} creatorInPage={false}/>}*/}
            {species === undefined && <ExistingSpeciesSelector initialSpecies={species} doSelect={s => {
                setSelectedSpecies(s)
            }} />}
            <Subform>
            {/* NAME (ID) */}
            <NameArea classNames={"inlineChildren"} currentName={name} headerTxt={"New Subspecies Name: "} setName={setName} readonly={false}/>
            </Subform>
            <Subform>
                {/* Aliases */}
            <AliasesArea readonly={false} updateParent={setAliases}/>
            </Subform>
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
    const errArea = () => {
        return <ErrorDisplay err={err}/>
    }
    const toggleButton = () => {
        return <div>
            <button className={"basicButton"} onClick={() => {
                setSelectorOpen(!selectorOpen)
            }}>{selectorOpen ? "Close subspecies selector" : (selected ? "Choose a different subspecies" : "Select a subspecies")}</button>
        </div>
    }
    if (!species){
        return null
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
            <ErrorDisplay err={err}/>
            <div>{"loading subspecies selector"}</div>
        </div>
    }
    if (subspeciesList.length == 0) {
        return <div className={"centerHChildren gapTop gapBottom"}>
            <ErrorDisplay err={"No Subspecies Found for species: " + (species && species)}/>
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

export function SubspeciesListPageTable({data, onClick, withLink}: ListPageItems<SubspeciesData>) {
    let cols: ListTableColumn<SubspeciesData>[] = [
        NewColumn("Species", (v)=>v.species, true),
        NewColumn("Subspecies", (v)=>v._id, true),
        NewColumn("Aliases", (v) => <div>
            {v.aliases && v.aliases.map((a, i) => {
                return <div key={a + i}>{a}</div>
            })}
        </div>, true), // TODO: ok?
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }), // TODO: fit?
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: SubspeciesData)=>{
            return <EntryLinkWrapper props={{entry:v,openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new SubspeciesData(v)}}/>
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