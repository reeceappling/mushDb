'use client'

import React, {JSX, useContext, useEffect, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {
    AddCreatedTriColFunction,
    AllEntries,
    Data,
    OnViewCreatorQuadCol
} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {
    IsValidNutrient,
    Nutrient,
    NutrientsAreaReadOnly,
    NutrientsEntriesGroupForNew
} from "@/app/components/formSubcomponents/nutrients";
import {
    IsValidSugar,
    Sugar,
    SugarEntriesGroupForNew,
    SugarsAreaReadOnly
} from "@/app/components/formSubcomponents/sugars";
import {
    CreatedLinkFor,
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest,
    DoUpdateRequest,
    ExistingDualSelector,
    FlexedArea,
    FlexedSinglesGroup,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    NewColumn,
    NewEntryFormWrapper,
    NewEntryInput,
    NumberToDateStr,
    OptionalArrayOfType,
    RequiredArrayOfType,
    RequiredKey,
} from "@/app/components/common";
import EntryLinkForId, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {
    Additive,
    AdditiveEntriesGroupForNew,
    AdditivesAreaReadOnly,
    IsValidAdditive
} from "@/app/components/formSubcomponents/additives";
import {JarRecipeData} from "@/app/components/jarRecipeServer";
import {ErrorDisplay, NameArea, StandardArea} from "@/app/components/formSubcomponents/commonClient";
import {Grain, GrainsEntriesGroupForNew, IsValidGrain} from "@/app/components/formSubcomponents/grains";
import {AclDisplay, MarshalAcl, TogglableAreaWithDepth, UnmarshalAcl} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import TestAndValidate from "@/app/components/testing/untested";
import {NewGrainBatchForm} from "@/app/components/grainBatchClient";
import {GrainBatchData} from "@/app/components/grainBatchServer";
import {OnViewCreatorsTriColArea} from "@/app/components/formSubcomponents/ovc";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {InputText} from "@/app/components/formSubcomponents/numericInput";


export function AssertJarRecipe(input: any): asserts input is JarRecipeData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['name', 'string'],
        ['standard', 'boolean'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Grain Jar Recipe assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // complex required keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl]
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('Jar Recipe assertion failure: required key ' + key + ' was not valid');
        }
    }

    // complex required array keys
    const complexRequiredArrayKeys = new Map<string, (v: any) => boolean>([
        ['grains', IsValidGrain], // TODO: ensure length > 0
    ])
    for (const [key, validator] of complexRequiredArrayKeys) {
        if (!RequiredArrayOfType(key, input, validator)) {
            throw new Error('Grain Jar Recipe assertion failure: required array key ' + key + ' was not valid');
        }
    }
    if (input['grains'].length < 1) {
        throw new Error('Grain Jar Recipe assertion failure: must have at least 1 grain type');
    }
    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['nutrients', IsValidNutrient],
        ['sugars', IsValidSugar],
        ['additives', IsValidAdditive],
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Grain Jar Recipe assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

function dataFormatFor<T>(init: T[] | undefined): Data<T>[] {
    return (init || []).map((v) => {
        return {data: v, disabled: false}
    })
}

export default function JarRecipeDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput<JarRecipeData>) {
    const [initial, setInitial] = useState(data)
    // grain non-changeable (base grain)
    // name non-changeable
    const [name, setName] = useState(initial.name)
    const [isStandard, setIsStandard] = useState(initial.standard)
    const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
    const [acl, setAcl] = useState<ACL>(initial.acl)
    const [err, setErr] = useState<string | undefined>()
    const updateInitial = (updated: JarRecipeData) => {
        setInitial(updated)
        setIsStandard(updated.standard)
        setNotes(InitialNotesState(updated.notes))
        setAcl(updated.acl)
        setErr(undefined)
    }
    const cookies = useContext(CookiesContext)
    const submit = () => {
        const body: any = {
            name: name,
            standard: isStandard,
            notes: notes,
            acl: MarshalAcl(acl),
        }
        DoUpdateRequest("jarRecipe", initial._id, body, AssertJarRecipe, allCookies(cookies))
            .then(v => {
                updateInitial(new JarRecipeData(v))
            })
            .catch(e => {
                setErr("failed to update initial: " + JSON.stringify(e))
            })
    }
    const jarGrainsArea = () => {
        return <div>
            {"Grains: "}
            {data.grains.map((g, i) => {
                return <div key={g.grain}>{g.percentage + "% " + g.grain}</div>
            })}
        </div>
    }
    const ovcs: OnViewCreatorQuadCol[] = [
        // TODO: NEW JAR FROM RECIPE (Also makes intermediary batch)?
        {
            txt: "New Batch From Recipe",
            newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                return <NewGrainBatchForm recipe={initial} handlers={{
                    onCreate: (newItem: GrainBatchData) => {
                        return onCreate([{
                            typeText: "Grain Batch",
                            node: <CreatedLinkFor linkId={newItem._id} typ={"grainBatch"}/>
                        }], false)
                    },
                    isTopLevel: false,
                }}/>
            },
        }
    ]
    return <DisplayFormWrapper entryType={"jarRecipe"}>
        <ErrorDisplay err={err} headerLevel={headerLevel}/>
        <ID props={{
            id: data._id,
            txt: "Grain Jar Recipe",
            entryType: "jarRecipe"
        }}>{/* TODO: modify top areas to dynamically size?*/}
            <NameModifiable initial={initial.name} readonly={readonly} updateParent={setName}/>
        </ID>
        <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>
        <FlexedArea>
            <FlexedSinglesGroup>
                <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
            </FlexedSinglesGroup>
            <FlexedSinglesGroup>
                <StandardArea isStandard={isStandard} readonly={readonly} setStandard={setIsStandard}
                              headerLevel={headerLevel}/>
            </FlexedSinglesGroup>
        </FlexedArea>
        {jarGrainsArea()}

        <NutrientsAreaReadOnly values={initial.nutrients}/>
        <SugarsAreaReadOnly values={initial.sugars}/>
        <AdditivesAreaReadOnly values={initial.additives}/>
        <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
        <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
            <AclDisplay initial={acl} readonly={readonly} updateParent={setAcl}/>
        </TogglableAreaWithDepth>
        {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
            e.stopPropagation();
            submit()
        }}>{"Update"}</button>}

    </DisplayFormWrapper>
}

export function NameModifiable({readonly, initial, updateParent}: {
    readonly: boolean,
    initial: string,
    updateParent: (name: string) => void
}) {
    const [name, setName] = useState(initial);
    const [modifying, setModifying] = useState(false);
    useEffect(() => {
        setName(initial)
        setModifying(false)
    }, [initial]);
    return <div className={"inlineChildren"}>
        {modifying ? <div><InputText value={name} readonly={false} errorMessage={"invalid name"} placeholder={"name"}
                                     onChange={s => {
                                         setName && setName(s || "")
                                         if (s) {
                                             updateParent(s)
                                         }
                                     }}
                                     onBlur={() => {
                                         setModifying(false)
                                     }}/></div> : <div onClick={e => {
            {/* TODO: onHover change color?*/
            }
            e.stopPropagation();
            if (!readonly) {
                setModifying(true)
            }
        }}>{name}</div>} {/*TODO: ensure type is called out as well?*/}
    </div>
}

export function NewJarRecipeForm({handlers}: { handlers: NewEntryInput<JarRecipeData> }) {
    const [name, setName] = useState<string | undefined>()
    const [grains, setGrains] = useState<Grain[]>()
    const [isStandard, setIsStandard] = useState(false)
    const [nutrients, setNutrients] = useState<Nutrient[]>([])
    const [sugars, setSugars] = useState<Sugar[]>([])
    const [additives, setAdditives] = useState<Additive[]>([])
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
    const [templateSelectorOpen, setTemplateSelectorOpen] = useState<boolean>(false)
    const loadTemplate = (template: JarRecipeData) => {
        setGrains(template.grains)
        setNutrients(template.nutrients || [])
        setSugars(template.sugars || [])
        setAdditives(template.additives || [])
        setTemplateSelectorOpen(false)
    }

    const cookies = useContext(CookiesContext)
    const newJarRecipeSubmit = () => {
        if (!name) {
            setErr("Name must be set!")
            return
        }
        if (!grains || grains.length < 1) {
            setErr("Grains must be set!")
            return
        }
        const gs = grains || []
        let totalPct = 0
        for (let i = 0; i < gs.length; i++) {
            if (gs[i].percentage < 0 || gs[i].percentage > 100) {
                setErr("Invalid grain percentage")
            }
            totalPct += gs[i].percentage
        }
        if (totalPct != 100) {
            setErr("Grain percentages must equal 100")
            return
        }
        const body: any = {
            name: name,
            grains: grains,
            standard: isStandard,
            nutrients: nutrients,
            sugars: sugars,
            additives: additives,
            notes: notes,
        }
        DoCreateRequest("jarRecipe", body, AssertJarRecipe, allCookies(cookies))
            .then(v => {
                handlers.onCreate ? handlers.onCreate(v) : console.log("no onCreate provided")
            })
            .catch(e => {
                setErr(JSON.stringify(e))
            })
    }
    const templateRecipeSelector = () => {
        if (templateSelectorOpen) {
            return <JarRecipeSelector doSelect={(rec) => { // TODO: endpoint for getStandard?
                if (rec === undefined) {
                    return
                }
                loadTemplate(rec)
            }} allowCreate={false}/>
        } else {
            return <button className={"basicButton"} onClick={() => {
                setTemplateSelectorOpen(true)
            }}>{"Select a template recipe (optional)"}</button>
        }
    }
    return <NewEntryFormWrapper entryType={"jarRecipe"}>
        <ErrorDisplay err={err}/>
        <TestAndValidate todos={["TEST THIS"]}>
            {templateRecipeSelector()}
        </TestAndValidate>
        <NameArea currentName={name} classNames={"inlineChildren"} readonly={false} setName={setName}/>

        {/* TODO: SUBFORMS!*/}
        <div>{"Grains"}</div>
        <GrainsEntriesGroupForNew currentEntries={grains || []} updateParent={setGrains}/>
        <StandardArea readonly={false} setStandard={setIsStandard}/>
        <div>{"Nutrients"}</div>
        <NutrientsEntriesGroupForNew currentEntries={nutrients} updateParent={setNutrients}/>
        <div>{"Sugars"}</div>
        <SugarEntriesGroupForNew currentEntries={sugars} updateParent={setSugars}/>
        <div>{"Additives"}</div>
        <AdditiveEntriesGroupForNew currentEntries={additives} updateParent={setAdditives}/>
        <NewEntryNotes setNotes={setNotes}/>
        <button className={"greenButton buttonFullWidth"} onClick={newJarRecipeSubmit}>{"Create Jar Recipe"}</button>
    </NewEntryFormWrapper>
}

export const JarRecipeArea = ({recipeId}: { recipeId?: string, headerLevel?: number }) => {
    let linkArea: JSX.Element | null = <div>{"unknown"}</div>
    if (recipeId !== undefined) {
        const b58id = recipeId
        linkArea = <EntryLinkForId
            props={{
                displayId: b58id,
                linkId: b58id,
                entryType: "jarRecipe",
                openInNewTab: false, // TODO: ok?
            }}/>// TODO: display name as well?
    }
    return <div>
        <div>{"Grain Recipe ID: "}</div>
        {linkArea}
    </div>
}

export function JarRecipeListPageTable({data, onClick, withLink}: ListPageItems<JarRecipeData>) {
    let cols: ListTableColumn<JarRecipeData>[] = [
        NewColumn("ID", (v) => v._id, true),
        NewColumn("Name", (v) => v.name, true), // TODO: shortname?
        NewColumn("Grains", (v) => {
            return <div>
                {v.grains.map((g, i) => {
                    return <div key={g.grain + i}>{g.grain}</div>
                })}
            </div>
        }, true),
        NewColumn("Nutrients", (v) => {
            return <div>
                {v.nutrients && v.nutrients.map((v, i) => {
                    return <div key={v.nutrient + i}>{v.nutrient}</div>
                })}
            </div>
        }, true),
        NewColumn("Sugars", (v) => {
            return <div>
                {v.sugars && v.sugars.map((v, i) => {
                    return <div key={v.type + i}>{v.type}</div>
                })}
            </div>
        }, true),
        NewColumn("Additives", (v) => {
            return <div>
                {v.additives && v.additives.map((v, i) => {
                    return <div key={v.additive + i}>{v.additive}</div>
                })}
            </div>
        }, true),
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }) // TODO: fit?
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: JarRecipeData) => {
            return <EntryLinkWrapper props={{entry: v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick} newClass={v => {
        return new JarRecipeData(v)
    }}/>
}

export function JarRecipeSelectorTable({data, onClick}: ListPageItems<JarRecipeData>) {
    const cols: ListTableColumn<JarRecipeData>[] = [
        NewColumn("Name", (v) => v.name), // TODO: shortname?
        NewColumn("ID", (v) => v._id),
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
        NewColumn("Link", (v: JarRecipeData) => {
            return <EntryLinkWrapper props={{entry: v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })
    ]
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick} newClass={v => {
        return new JarRecipeData(v)
    }}/>
}

export function JarRecipeSelector(
    {
        doSelect,
        allowCreate,
        creatorInPage,
    }: {
        doSelect: (val: JarRecipeData | undefined) => void,
        allowCreate?: boolean,
        creatorInPage?: boolean
    }) {
    const table = (items: JarRecipeData[]): JSX.Element => {
        return <JarRecipeSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingDualSelector entryType={"jarRecipe"} entryTypes={"jarRecipes"} doSelect={doSelect}
                                 asserter={AssertJarRecipe}
                                 table={table}>
        {allowCreate && (creatorInPage ? <NewJarRecipeForm handlers={{onCreate: doSelect, isTopLevel: false}}/> :
            <TestAndValidate todos={["fix creator"]}>
                <div>{"LINK TO CREATOR HERE FIXME"}</div>
            </TestAndValidate>)}
    </ExistingDualSelector>
}