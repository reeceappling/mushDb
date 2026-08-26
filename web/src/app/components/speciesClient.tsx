'use client'

import {
    IsValidNote,
    NewEntryNotes,
    Note,
    NotesFormArea
} from "@/app/components/formSubcomponents/notes";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import React, {JSX, useContext, useEffect, useState} from "react";
import {
    AddCreatedTriColFunction,
    AllEntries,
    OnViewCreatorTriCol
} from "@/app/components/formSubcomponents/shared";
import {SpeciesData, SpeciesSelectorCloseable} from "@/app/components/speciesServer";
import {
    clientPostRequestHeaders,
    CreateNewEntryButton,
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest,
    DoUpdateRequest,
    ErrHandler,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    HandleJsonResponse,
    IsString,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    NewColumn,
    NewEntryFormWrapper,
    NewEntryInput,
    NumberToDateStr,
    OptionalArrayOfType,
    RequiredKey,
    SelectorWrapper,
    Subform
} from "@/app/components/common";
import {AliasesArea, ErrorDisplay, NameArea} from "@/app/components/formSubcomponents/commonClient";
import {SubstrateRecipeArea, SubstrateRecipeSelector} from "@/app/components/substrateRecipeClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {
    AclDefaultAclDisplay,
    AclDisplay,
    MarshalAcl,
    NewAllCanWriteAcl,
    UnmarshalAcl
} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {SubstrateRecipeData} from "@/app/components/substrateRecipeServer";
import EntryLinkForId, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {CreatedLinkTriCol, OnViewCreatorsTriColArea} from "@/app/components/formSubcomponents/ovc";
import {NewSubspeciesForm} from "@/app/components/subspeciesClient";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {SelectorFor} from "@/app/components/selector";
import {ActionTypes, useModalContext} from "@/app/components/formSubcomponents/modalContext/modal";
import {JarRecipeData} from "@/app/components/jarRecipeServer";

export function AssertSpecies(input: any): asserts input is SpeciesData {
    if (typeof input !== 'object') {
        throw 'Input is not an object! Input is ' + typeof input
    }
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['scientificName', 'string'],
        ['standardSubstrate', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw 'Species assertion failure: ' + key + ' was not type ' + expType + '. Was ' + (typeof input[key]);
        }
    }
    // complex required keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl],
        //['defaultAcl', IsValidAcl]
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw 'Species assertion failure: required key ' + key + ' was not valid'
        }
    }
    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
        ['aliases', IsString],
        ['subspecies', IsString],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw 'Species assertion failure: optional array key ' + key + ' was not valid'
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

export default function SpeciesDisplay(
    {
        readonly, data, headerLevel
    }: DisplayInput<SpeciesData>) {
    const {dispatch} = useModalContext();
    const [initial, setInitial] = useState(data)

    const [substrate, setSubstrate] = useState<string | undefined>(data.standardSubstrate)
    const [subspecies, setSubspecies] = useState<string[] | undefined>(data.subspecies)
    const [aliases, setAliases] = useState<string[]>(initial.aliases || [])
    const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
    const [err, setErr] = useState<string | undefined>(undefined)
    const [acl, setAcl] = useState<ACL>(initial.acl)
    const [defaultAcl, setDefaultAcl] = useState<ACL>(initial.defaultAcl)
    const updateInitial = (updated: SpeciesData) => {
        setInitial(updated)
        setSubstrate(updated.standardSubstrate)
        setSubspecies(updated.subspecies)
        setAliases(updated.aliases || [])
        setNotes(InitialNotesState(updated.notes))
        setAcl(updated.acl)
        setDefaultAcl(updated.defaultAcl)
        setErr(undefined)
    }
    const cookies = useContext(CookiesContext)
    const update = () => {
        // Notes, aliases, substrate recipe, and have only
        const body: any = {
            substrate: substrate, // TODO: something is going wrong with this? validate serverside
            notes: notes,
            aliases: aliases,
            acl: MarshalAcl(acl),
            defaultAcl: MarshalAcl(defaultAcl),
        }
        DoUpdateRequest("species", encodeURIComponent(data._id), body, AssertSpecies, allCookies(cookies))
            .then(v => {
                updateInitial(new SpeciesData(v))
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Update Success",
                        text: "entry updated successfully",
                        isErr: false
                    }})
            })
            .catch(e => {
                setErr("failed to update initial: " + JSON.stringify(e))
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Update Failed",
                        text: "failed to update: " + JSON.stringify(e),
                        isErr: true
                    }})
            })
    }
    const ovcs: OnViewCreatorTriCol[] = [
        {
            txt: "Create New Subspecies",
            newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                return <NewSubspeciesForm species={initial} handlers={{
                    onCreate: (v: SubspeciesData) => {
                        setSubspecies([...(subspecies || []), v._id])
                        const toAdd: CreatedLinkTriCol[] = []
                        // const toAdd = [{ // TODO: FIX?
                        //     typeText: "Subspecies",
                        //     node: <CreatedLinkFor
                        //         linkText={v._id}
                        //         linkId={encodeURI(v._id)}
                        //         typ={"subspecies"}/>
                        // }]
                        return onCreate(toAdd, true)
                    },
                    isTopLevel: false,
                }}/>
            },
        },
    ]
    const updateAliases = (as: string[])=>{
        console.log(JSON.stringify(as)) // TODO: del
        setAliases(as)
    }
    return (
        <DisplayFormWrapper entryType={"species"}>
            <ErrorDisplay err={err}/>
            <ID props={{id: data._id, txt: "Species", entryType: "species"}}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <div>{"Scientific Name: " + initial.scientificName}</div>
                    <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <SubstrateRecipeArea txt={"Standard Substrate Recipe: "} id={substrate} readonly={false} onSelect={s => {
                            s && setSubstrate(s._id)
                        }}/>
                </FlexedSinglesGroup>
            </FlexedArea>
            <AliasesArea initial={initial.aliases} readonly={readonly} updateParent={updateAliases}/>{/* TODO: initial as just aliases?*/}
            <SubspeciesForSpeciesArea subspecies={subspecies}/>
            <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>
            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            <AclDefaultAclDisplay ACL={initial.acl} defaultACL={initial.defaultAcl} updateAcl={setAcl} updateDefaultAcl={setDefaultAcl}
                                  readonly={readonly}/>
            {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                e.stopPropagation();
                update()
            }}>{"Update"}</button>}
        </DisplayFormWrapper>
    )
}

export function NewSpeciesForm(
    {handlers, substrateIn}:
    { handlers: NewEntryInput<SpeciesData>, substrateIn?: SubstrateRecipeData }
) {
    const {dispatch} = useModalContext();
    const [name, setName] = useState("")
    const [sciName, setSciName] = useState("")
    const [aliases, setAliases] = useState<string[]>([])
    const [sub, setSub] = useState(substrateIn)
    const [notes, setNotes] = useState<Note[]>([])
    const [acl, setAcl] = useState<ACL>(NewAllCanWriteAcl())
    const [err, setErr] = useState<string | undefined>();

    const cookies = useContext(CookiesContext)
    const baseAcl = NewAllCanWriteAcl()
    const submitNewSpecies = () => {
        console.log("submitting new species")
        if (name === "") {
            setErr("Name must not be blank!")
            return
        }
        if (sub === undefined) {
            setErr("Substrate must not be blank!")
            return
        }
        const body: any = {
            name: name,
            scientificName: sciName,
            aliases: aliases,
            recipe: sub._id, // substrate recipe
            notes: notes,
            acl: MarshalAcl(acl),
            // defaultAcl starts same as ACL...
        }
        DoCreateRequest("species", body, AssertSpecies, allCookies(cookies))
            .then(v => {
                if(handlers.onCreate!==undefined){
                    handlers.onCreate(new SpeciesData(v))
                    handlers.isTopLevel && dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                            header: "Create Success",
                            text: "entry created successfully",
                            isErr: false
                        }})
                } else {
                    console.log("no onCreate provided")
                }
            })
            .catch(e => {
                setErr(JSON.stringify(e))
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Create Failure",
                        text: "entry failed to create: " + JSON.stringify(e),
                        isErr: true
                    }})
            })
    }
    return <NewEntryFormWrapper entryType={"species"} isTopLevel={handlers.isTopLevel}>
        <ErrorDisplay err={err}/>
        <Subform>
        <NameArea classNames={"inlineChildren"} currentName={name} headerTxt={"Name :"} setName={setName}/>
        <NameArea classNames={"inlineChildren"} currentName={sciName} headerTxt={"Scientific Name :"}
                  setName={setSciName}/>
        </Subform>
        <Subform>
        <AliasesArea updateParent={setAliases} readonly={false}/>
        </Subform>
        <Subform>
        <SelectorWrapper current={sub} title={"Standard Substrate"} nameFunc={(v: SubstrateRecipeData) => v._id}>
            <SubstrateRecipeSelector doSelect={setSub} allowCreate={handlers.isTopLevel} creatorInPage={false}/>
        </SelectorWrapper>
        </Subform>
        <NewEntryNotes setNotes={setNotes}/>
        <AclDisplay readonly={false} initial={baseAcl} updateParent={setAcl}/>
        {/* SUBMIT AREA */}
        <CreateNewEntryButton onSubmit={submitNewSpecies}/>
    </NewEntryFormWrapper>
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
    //const cookies = useContext(CookiesContext)
    //const [expandedAfterSelected, setExpandedAfterSelected] = useState<boolean>(false)
    const [isLoaded, setLoaded] = useState(false)
    const [speciesList, setSpeciesList] = useState<SpeciesData[]>([]); // TODO: add subspecies to species data?!!!!!
    const [selectorOpen, setSelectorOpen] = useState(false)
    // TODO: REFRESH WHEN NEEDED????
    const [selected, setSelected] = useState<SpeciesData | undefined>()
    const [err, setErr] = useState<string | undefined>(undefined)
    useEffect(() => {
        fetch(BaseExternalUrl + "/db/list/species", {
            method: "GET",
            headers: clientPostRequestHeaders,
        })
            .then(HandleJsonResponse)
            .then((data) => {
                if (!Array.isArray(data)) {
                    setErr("Server response is not an array")
                }
                data = data as any[]
                for (let i = 0; i < data.length; i++) {
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
    if (!selected && !selectorOpen) {
        return <div className={"centerHChildren gapTop"}>
            <ErrorDisplay err={err}/>
            <div>
                <button className={"basicButton"} onClick={() => {
                    setSelectorOpen(true)
                }}>{"Select a species"}</button>
            </div>
        </div>
    }
    if (!isLoaded) {
        return <div className={"centerHChildren gapTop"}>
            <ErrorDisplay err={err}/>
            <div>{"loading species selector"}</div>
        </div>
    }
    if (speciesList.length == 0) {
        return <div className={"centerHChildren gapTop"}>
            <ErrorDisplay err={"No Species Found on server"}/>
            {/* TODO: CREATE SPECIES BUTTON*/}
            <div>{"CREATE SPECIES LINK"}</div>
        </div>

    }

    if (selected && !selectorOpen) {
        return <div>
            <ErrorDisplay err={err}/>
            <div>
                {"Currently Selected species: "/* TODO: OVERHAUL*/}
                <div>{selected._id}</div>
                <button className={"basicButtonSmall"} onClick={() => {
                    setSelectorOpen(true)
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
            <SpeciesSelector doSelect={s => {
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
        NewColumn("Name", (v) => v._id, true),
        NewColumn("Scientific", (v) => v.scientificName, true), // TODO: wrap?
        NewColumn("Aliases", (v) => <div>
            {v.aliases && v.aliases.map((a, i) => {
                return <div key={a + i}>{a}</div>
            })}
        </div>, true),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }), // TODO: fit?
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: SpeciesData) => {
            return <EntryLinkWrapper props={{entry: v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v => {
        return new SpeciesData(v)
    }}/>
}

export function SpeciesSelectorTable({data, onClick}: ListPageItems<SpeciesData>) {
    return <SpeciesListPageTable data={data} onClick={onClick} withLink={true}/>
}

export function SpeciesSelector(
    {
        doSelect
    }: {
        doSelect: (val: SpeciesData | undefined) => void,
    }) {
    const table = (items: SpeciesData[]): JSX.Element => {
        return <SpeciesSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"species"} entryTypes={"species"} doSelect={doSelect}
                                   asserter={AssertSpecies}
                                   table={table}>
    </ExistingRecentSelector>
}

export function SubspeciesForSpeciesArea(
    {
        subspecies
    }: {
        subspecies?: string[]
    }) {
    if (!subspecies) {
        return null
    }
    return <div>{/* TODO: depth? */}
        <div className={"text-md"/* TODO: OK? */}>{"Subspecies :"}</div>
        {subspecies.map((subsp, i) => {
            return <div key={subsp}>
                <EntryLinkForId props={{
                    entryType: "subspecies",
                    linkId: encodeURI(subsp),
                    displayId: subsp,
                    openInNewTab: false, // TODO: ok?
                }}/>
            </div>
        })}
    </div>
}

export function ExistingSpeciesSubspeciesSelector(
    {
        doSelectSpecies,
        doSelectSubspecies,
        initialSpecies,
    }: {
        doSelectSpecies: (val?: SpeciesData) => void,
        doSelectSubspecies: (val?: string) => void,
        headerLevel?: number,
        initialSpecies?: SpeciesData,
    }) {
    const [subspeciesOptions, setSubspeciesOptions] = useState<string[] | undefined>(initialSpecies?.subspecies)
    return <div>
        {initialSpecies === undefined && <SpeciesSelectorCloseable doSelect={(sp) => {
            if (!sp) {
                setSubspeciesOptions(undefined)
                doSelectSubspecies(undefined)
                doSelectSpecies(undefined)
            } else {
                doSelectSpecies(sp)
                setSubspeciesOptions(sp.subspecies)
            }
        }} txt={"Species"} allowCreation={false}/>}
        {(subspeciesOptions && subspeciesOptions.length > 0) && <div className={"inlineChildren"}>
            <div>{"Subspecies: "}</div>
            <div>
                <SelectorFor initial={""} options={["", ...subspeciesOptions]} updateParent={subs => {
                    if (subs === "") {
                        doSelectSubspecies(undefined)
                    } else {
                        doSelectSubspecies(subs)
                    }
                }} disabled={false}/>
            </div>
        </div>}
    </div>
}