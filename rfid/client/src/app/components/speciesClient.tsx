'use client'

import NotesAreaOld, {
    IsValidNote, NewEntryNotes,
    Note,
    NoteEntriesGroup,
    NotesAreaInline
} from "@/app/components/formSubcomponents/notes";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import React, {useEffect, useState} from "react";
import {AllEntries, Data} from "@/app/components/formSubcomponents/shared";
import {SpeciesData} from "@/app/components/speciesServer";
import {
    CreateNewEntryButton,
    DisplayInput, HandleJsonResponse,
    InlineExpansionArea, InlineExpansionButton,
    InlineProps,
    InlineSubArea, IsString, NewEntryInput,
    OptionalArrayOfType, OptionalKey, SingleListProps
} from "@/app/components/common";
import {AliasesArea, ErrorDisplay, NameArea} from "@/app/components/formSubcomponents/commonClient";
import {SubstrateRecipeArea, SubstrateRecipeSelector} from "@/app/components/substrateRecipeClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {
    AclDisplay,
    DefaultAclDisplay,
    IsValidAcl,
    TogglableAreaWithDepth
} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import TestAndValidate from "@/app/components/testing/untested";
import {SubstrateRecipeData} from "@/app/components/substrateRecipeServer";
import {DisplayFormWrapper, NewEntryFormWrapper} from "./lcRecipeClient";
import {DepthProvider} from "@/app/components/formSubcomponents/depthContext/depth";
import {FlexedArea, FlexedSinglesGroup, NotesFormArea} from "@/app/components/agarBatchClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";

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

        const [substrate, setSubstrate] = useState(data.standardSubstrate)
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
        const update = ()=>{
            // Notes, aliases, substrate recipe, and have only

            fetch(BaseExternalUrl+"/db/update/species/"+encodeURI(data._id), { // TODO: ensure correct
                method: "POST",
                headers: {
                    credentials: 'include',
                    //'Cookie': cookies,
                    'Content-type': 'application/json'
                },
                body: JSON.stringify({
                    substrate: substrate, // TODO: something is going wrong with this
                    notes: notes,
                    aliases:aliases,
                    acl: acl,
                    defaultAcl: defaultAcl,
                })
            })
                .then(HandleJsonResponse)
                .then((entry) => {
                    AssertSpecies(entry)
                    updateInitial
                })
                .catch((error) => {
                    setErr(JSON.stringify(error))
                });
        }
        return ( // TODO: FIX THIS WHOLE FUNC!
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
                <TestAndValidate todos={["ONLY ONE ACL, ONE DEFAULT ACL","ENSURE MODIFYING ACLS WORKS AS NEEDED"]}>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl} />
                </TogglableAreaWithDepth>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view default ACL"} closeTxt={"minimize default ACL area"}>
                    <DefaultAclDisplay readonly={readonly} ACL={defaultAcl} updateParent={setDefaultAcl}/>
                </TogglableAreaWithDepth>
                </TestAndValidate>

                {readonly ? null : <input type="submit" value="Update" onClick={update} onSubmit={(e)=>{e.preventDefault();}}/>}
                {/* TODO: ? <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>*/}
            </DisplayFormWrapper>
        )
    } catch (err){
        return <div>{"ERROR: Species data format incorrect: "+err}</div> // TODO: CHANGE?????
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
    const [err, setErr] = useState<string | undefined>()
    //const [perms, setPerms] = useState<EntryPerms | undefined>()
    ////const [cookies, setCookie, removeCookie] = useCookies(['SessionId']);
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
        fetch(BaseExternalUrl+"/db/create/species", {
            method: 'Post',
            body: JSON.stringify({
                name:name,
                scientificName:sciName,
                aliases:aliases,
                sub:sub._id,
                notes:notes,
            }),
            headers: {
                credentials: 'include',
                'Content-type': "application/json"
            },
        })
            .then(HandleJsonResponse)
            .then(entry=> {
                AssertSpecies(entry)
                handlers.onCreate && handlers.onCreate(entry)
            })
            .catch((err) => {
                setErr(JSON.stringify(err))
            });
    }
    return (
        <NewEntryFormWrapper entryType={"species"}>
            <NameArea classNames={"inlineChildren"} currentName={name} headerTxt={"Name :"} setName={setName}/>
            <NameArea classNames={"inlineChildren"} currentName={sciName} headerTxt={"Scientific Name :"} setName={setSciName}/>
            <AliasesArea aliases={aliases} updateParent={setAliases} readonly={false}/> {/* TODO: OVERHAUL */}
            {substrateIn === undefined && <SubstrateRecipeSelector txt={"Standard Substrate: "} doSelect={setSub} allowCreation={handlers.isTopLevel} creatorInPage={false/* TODO: isTopLevel and embedded creators*/}/>}
            <NewEntryNotes setNotes={setNotes}/>
            {/* SUBMIT AREA */}
            <CreateNewEntryButton onSubmit={submitNewSpecies}/>
        </NewEntryFormWrapper>
    )
}


export function SpeciesInline({data, expandByDefault, onClick, showMainPageButton, idIsLink}: InlineProps<SpeciesData>) {
    const [expanded, setExpanded] = useState(expandByDefault)
    return <div className={"fullWidth"}>
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
    </div>
}

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
        //cookies: string,
    }) {
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

        fetch(BaseExternalUrl + "/db/list/species", {
            method: "GET",
            headers: {
                credentials: 'include',
                //'Cookie': cookies,
                // TODO: THIS!
            },
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
            .catch((error) => {
                setErr(JSON.stringify(error))
            });
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
                {"Currently Selected species: "}{expandedAfterSelected?<div><SpeciesInline data={selected} headerLevel={headerLevel}/><button className={"basicButton"} onClick={()=>{setExpandedAfterSelected(false)}}>Show ID only</button></div>:<div>{selected._id}<button onClick={()=>{setExpandedAfterSelected(true)}}>Show More</button></div>}
                <button onClick={() => {
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
        <DepthProvider>{/* TODO: is this necessary here?*/}
        {closeButton}{/* TODO: THIS DOES NOT WORK */}
        {speciesList.map((spec, i) => {
            return <div key={i} className={"gapTop"}>
                <SpeciesInline key={i} data={spec} headerLevel={headerLevel} onClick={sp => { // TODO: FIX THIS SO ITS ACTUALLY INLINE!
                    console.log("selected: "+(sp?._id || "undefined")) // TODO: del
                    doSelect(sp)
                    setSelectorOpen(false)
                    setSelected(sp)
                }}/>
            </div>
        })}
        {closeButton}
        </DepthProvider>
    </div>
}
