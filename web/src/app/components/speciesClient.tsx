'use client'

import {
    IsValidNote, NewEntryNotes,
    Note,
    NotesAreaInline, NotesFormArea
} from "@/app/components/formSubcomponents/notes";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import React, {JSX, useContext, useEffect, useState} from "react";
import {AllEntries} from "@/app/components/formSubcomponents/shared";
import {SpeciesData} from "@/app/components/speciesServer";
import {
    clientPostRequestHeaders,
    createApiUrlFor,
    CreateNewEntryButton,
    DisplayFormWrapper,
    DisplayInput, DoCreateRequest, DoUpdateRequest, ErrHandler, ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    HandleJsonResponse,
    InlineEntry,
    InlineExpansionArea,
    InlineExpansionButton,
    InlineProps,
    InlineSubArea,
    IsString,
    ListPageItems, ListPageTable,
    ListTableColumn,
    NewColumn,
    NewEntryFormWrapper,
    NewEntryInput,
    NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey,
    SelectorWrapper,
    Subform, updateApiUrlFor
} from "@/app/components/common";
import {AliasesArea, ErrorDisplay, NameArea} from "@/app/components/formSubcomponents/commonClient";
import {SubstrateRecipeArea, SubstrateRecipeSelector} from "@/app/components/substrateRecipeClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {
    AclDefaultAclDisplay,
    IsValidAcl, MarshalAcl
} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import TestAndValidate from "@/app/components/testing/untested";
import {SubstrateRecipeData} from "@/app/components/substrateRecipeServer";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {AssertSlant} from "@/app/components/slantClient";
import {AssertProject} from "@/app/components/projectClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {AgarRecipeData} from "@/app/components/agarRecipeServer";

// TODO: list page not working

export function AssertSpecies(input: any): asserts input is SpeciesData {
    if (typeof input !== 'object') {
        throw 'Input is not an object! Input is ' + typeof input
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['scientificName', 'string'],
        ['standardSubstrate', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw 'Species assertion failure: ' + key + ' was not type ' + expType + '. Was ' + (typeof input[key]);
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
       ['acl', IsValidAcl],
        ['defaultAcl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw 'Plate assertion failure: optional key ' + key + ' was not valid' // TODO: change all throws to strings instead of errors
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
        ['aliases', IsString],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw 'Species assertion failure: optional array key ' + key + ' was not valid'
        }
    }
    return
}

export default function SpeciesDisplay(
    {
        id, readonly, data, headerLevel
    }: DisplayInput) {
    try {

        AssertSpecies(data)
        const [initial, setInitial] = useState(data)

        const [substrate, setSubstrate] = useState<string | undefined>(data.standardSubstrate)
        const [aliases, setAliases] = useState<string[]>(initial.aliases || [])
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [err, setErr] = useState<string | undefined>(undefined)
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const [defaultAcl, setDefaultAcl] = useState<ACL | undefined>(initial.defaultAcl)
        const updateInitial = (updated: SpeciesData): void => {
            setInitial(updated)
            setSubstrate(updated.standardSubstrate)
            setAliases(updated.aliases || [])
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
            setDefaultAcl(updated.defaultAcl)
        }
        const cookies = useContext(CookiesContext)
        const update = ()=>{
            // TODO: validate substrate?
            // Notes, aliases, substrate recipe, and have only
            const body: any = {
                substrate: substrate, // TODO: something is going wrong with this
                notes: notes,
                aliases:aliases,
                acl: MarshalAcl(acl), // TODO: ensure ok
                defaultAcl: MarshalAcl(defaultAcl), // TODO: ensure ok
            }
            DoUpdateRequest("species",encodeURIComponent(data._id), body, AssertSpecies, allCookies(cookies))
                .then(v=>{
                    updateInitial(new SpeciesData(v))
                })
                .catch(e=>{
                    setErr(JSON.stringify(e))
                })
        }
        return (
            <DisplayFormWrapper entryType={"species"}>
                <ErrorDisplay err={err} headerLevel={headerLevel} />
                <ID id={data._id} txt={"Species"} entryType={"species"}/>
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <div>{"Scientific Name: "+initial.scientificName}</div>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <TestAndValidate todos={["standardSubstrate is getting changed when it shouldnt (going to 1)"]}>
                            <SubstrateRecipeArea txt={"Standard Substrate Recipe: "} id={substrate} headerLevel={headerLevel} readonly={false} onSelect={s=>{s && setSubstrate(s._id)}}/>
                        </TestAndValidate>
                    </FlexedSinglesGroup>
                </FlexedArea>
                <AliasesArea aliases={aliases} readonly={readonly} updateParent={setAliases} headerLevel={headerLevel}/>
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <AclDefaultAclDisplay ACL={acl} defaultACL={defaultAcl} updateAcl={setAcl} updateDefaultAcl={setDefaultAcl} readonly={readonly}/>
                <button className={"basicButtonSmall"} onClick={(e)=>{e.stopPropagation();}} >
                    {"Load subspecies"}{/* TODO: view subspecies???? */}
                </button>

                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                    e.stopPropagation();
                    update()
                }}>{"Update"}</button>}
                <button className={"basicButtonSmall"} onClick={(e)=>{
                    e.stopPropagation();
                }} >
                    {"Create New Subspecies"}{/* TODO: create subspecies???? IN OVC AREA! */}
                </button>
            </DisplayFormWrapper>
        )
    } catch (err){
        return <div>{"ERROR: Species data format incorrect: "+err}</div>
    }
}

export function NewSpeciesForm(
    {handlers, substrateIn}:
    {handlers: NewEntryInput<SpeciesData>, substrateIn?: SubstrateRecipeData}
    ) {
    const [name, setName] = useState("")
    const [sciName, setSciName] = useState("")
    const [aliases, setAliases] = useState<string[]>([])
    const [sub, setSub] = useState(substrateIn)
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>();
    const errHandler = ErrHandler(setErr)
    const cookies = useContext(CookiesContext)
    const submitNewSpecies = () => {
        console.log("submitting new species")
        if(name===""){
            setErr("Name must not be blank!")
            return
        }
        if(sub===undefined){
            setErr("Substrate must not be blank!")
            return
        }
        const body: any = {
            name:name,
            scientificName:sciName,
            aliases:aliases,
            sub:sub._id,
            notes:notes,
        }
        DoCreateRequest("species", body, AssertSpecies, allCookies(cookies))
            .then(v=>{
                handlers.onCreate ? handlers.onCreate(v) : console.log("no onCreate provided")
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }
    return <NewEntryFormWrapper entryType={"species"}>
            <NameArea classNames={"inlineChildren"} currentName={name} headerTxt={"Name :"} setName={setName}/>
            <NameArea classNames={"inlineChildren"} currentName={sciName} headerTxt={"Scientific Name :"} setName={setSciName}/>
            <ErrorDisplay err={err}/>
            <AliasesArea aliases={aliases} updateParent={setAliases} readonly={false}/> {/* TODO: OVERHAUL */}
            <SelectorWrapper current={sub} title={"Standard Substrate"} nameFunc={(v: SubstrateRecipeData) => v._id}>
                <SubstrateRecipeSelector doSelect={setSub} allowCreate={handlers.isTopLevel} creatorInPage={false}/>
            </SelectorWrapper>
            <NewEntryNotes setNotes={setNotes}/>
            {/* SUBMIT AREA */}
            <CreateNewEntryButton onSubmit={submitNewSpecies}/>
        </NewEntryFormWrapper>
}


// TODO: get rid of?
export function SpeciesInline({data, expandByDefault, onClick, showMainPageButton, idIsLink}: InlineProps<SpeciesData>) {
    const [expanded, setExpanded] = useState(expandByDefault)
    return <InlineEntry  onClick={onClick}>
        <InlineSubArea props={{}}>
            <TestAndValidate todos={["BOLD THIS SO THAT WE KNOW TO CLICK IT"]}>
                <ID id={data._id} txt={"Species"} entryType={"species"} allowOpenMainPage={showMainPageButton} linkPage={idIsLink}/>
            </TestAndValidate>
            <NameArea headerTxt={"Scientific Name: "} readonly={true} currentName={data.scientificName}/>
            <AliasesArea aliases={data.aliases} readonly={true} />
            <SubstrateRecipeArea id={data.standardSubstrate} readonly={true}/>
        </InlineSubArea>
        <InlineExpansionArea props={{expanded: expanded}}>
            <NotesAreaInline notes={data.notes}  offset={-1}/>
            <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true} />
        </InlineExpansionArea><InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
                               expanded={expanded}/>
    </InlineEntry>
}

export function SpeciesSubspeciesArea({species, subspecies}: {
    subspecies?: string,
    species?: string,
}) {
    return <>
        <div>
            {"Species: " + (species || "")}{/* TODO: LINK!?*/}
        </div>
        <div>
            {"Subspecies: " + (subspecies || "")}{/* TODO: LINK!*/}
        </div>
    </>
}

// TODO: Distinguish from SpeciesSelector
export function ExistingSpeciesSelector(
    {
        doSelect,
        headerLevel,
        initialSpecies,
        //cookies,
    }: {
        doSelect: (val?: SpeciesData) => void,
        headerLevel?: number,
        initialSpecies?: string,
    }) {
    const cookies = useContext(CookiesContext)
    const [expandedAfterSelected, setExpandedAfterSelected] = useState<boolean>(false)
    const [isLoaded, setLoaded] = useState(false)
    const [speciesList, setSpeciesList] = useState<SpeciesData[]>([]);
    const [selectorOpen, setSelectorOpen] = useState(false)
    // TODO: REFRESH WHEN NEEDED????
    const [selected, setSelected] = useState<SpeciesData | undefined>()
    const [err, setErr] = useState<string | undefined>(undefined)
    ////const [cookies, setCookie, removeCookie] = useCookies(['SessionId']);
    // const speciesFor: (s: string) => SpeciesData = (s: string) => { // TODO: DELETEME
    //     return {
    //         _id: s,
    //         scientificName: s + "_scientific_name",
    //         have: true,
    //         standardSubstrate: 'substrate ID',
    //         lastUpdated: 0,
    //         //perms: {userPerms: {ids:[],canWrite:[]}, projectPerms: {ids:[],canWrite:[]}, blanketPerms: 2}
    //     }
    // }
    useEffect(() => {
        //setSpeciesList([speciesFor('spec A'), speciesFor('spec B')]) // TODO: REMOVE
        // setSelected(undefined)
        // setLoaded(true) // TODO: REMOVE
        // return // TODO: REMOVE

        fetch(BaseExternalUrl + "/db/list/species", { // TODO: simplify?
            method: "GET",
            headers: clientPostRequestHeaders,
        })
            .then(HandleJsonResponse)
            .then((data) => {
                if(!Array.isArray(data)){
                    setErr("Server response is not an array")
                }
                data = data as any[]
                for(let i=0; i<data.length;i++){
                    AssertSpecies(data[i])
                }
                setSpeciesList(data as SpeciesData[])
                if (initialSpecies) {
                    for (let i = 0; i < data.length; i++) {
                        if (data[i]._id == initialSpecies) {
                            setSelected(data[i])
                            break
                        }
                    }
                }
                setErr(undefined)
                setLoaded(true)
            })
            .catch(ErrHandler(setErr));
    }, []);
    // TODO: CLEAR SELECTION
    if(!selected && !selectorOpen) {
        return <div className={"centerHChildren gapTop"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <div>
                <button className={"basicButton"} onClick={() => {
                    setSelectorOpen(true)
                }}>{"Select a species"}</button>
            </div>
        </div>
    }
    if (!isLoaded) {
        return <div className={"centerHChildren gapTop"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <div>{"loading species selector"}</div>
        </div>
    }
    if (speciesList.length == 0) {
        return <div className={"centerHChildren gapTop"}>
            <ErrorDisplay err={"No Species Found on server"} headerLevel={headerLevel}/>
            {/* TODO: CREATE SPECIES BUTTON*/}
            <div>{"CREATE SPECIES LINK"}</div>
        </div>

    }

    if (selected && !selectorOpen) {
        return <div>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <div>
                {"Currently Selected species: "/* TODO: OVERHAUL*/}{expandedAfterSelected?<div><SpeciesInline data={selected} headerLevel={headerLevel}/><button className={"basicButtonSmall"} onClick={()=>{setExpandedAfterSelected(false)}}>Show ID only</button></div>:<div>{selected._id}<button className={"basicButtonSmall"} onClick={()=>{setExpandedAfterSelected(true)}}>Show More</button></div>}
                <button className={"basicButtonSmall"} onClick={() => {
                    setSelectorOpen(true)
                    setExpandedAfterSelected(false)
                }}>{"Choose a different species"}</button>
            </div>
        </div>
    }
    const closeButton = <button onClick={(e) => {
        e.preventDefault();
        setSelectorOpen(false)
    }}>{"Close Species Selector"}</button>
    return <div className={"gapBottom"}>
        <Subform>{/* TODO: is this necessary here?*/}
        {closeButton}{/* TODO: THIS DOES NOT WORK */}
            <SpeciesSelector doSelect={s=>{
                doSelect(s)
                setSelected(s)
                setSelectorOpen(false)
            }}/>
        {closeButton}
        </Subform>
    </div>
}

export function SpeciesListPageTable({data, onClick, withLink}: ListPageItems<SpeciesData>) {
    let cols: ListTableColumn<SpeciesData>[] = [
        NewColumn("Name", (v)=>v._id),
        NewColumn("Scientific", (v)=>v.scientificName),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: SpeciesData)=>{
            return <EntryLinkWrapper props={{entry:v,openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new SpeciesData(v)}}/>
}
export function SpeciesSelectorTable({data, onClick}: ListPageItems<SpeciesData>) {
    return <SpeciesListPageTable data={data} onClick={onClick} withLink={true} />
}
export function SpeciesSelector(
    {
        doSelect
    }: {
        doSelect: (val: SpeciesData | undefined) => void,
    }) {
    const table = (items: SpeciesData[]):JSX.Element=>{
        return <SpeciesSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"species"} entryTypes={"species"} doSelect={doSelect} asserter={AssertSpecies}
                                   table={table}>
    </ExistingRecentSelector>
}
